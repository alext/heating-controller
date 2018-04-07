package scheduler

import (
	"errors"
	"log"
	"sync"
	"time"
)

// variable indirection to enable testing
var time_Now = time.Now

var ErrInvalidJob = errors.New("invalid job")

//go:generate counterfeiter . Scheduler
type Scheduler interface {
	Start()
	Stop()
	Running() bool
	AddJob(Job) error
	RemoveJob(Job)
	NextJob() *Job
	ReadJobs() []Job
	Override(Job)
	CancelOverride()
}

type scheduler struct {
	id        string
	jobs      []*Job
	running   bool
	lock      sync.Mutex
	commandCh chan func()

	nextJob *Job
	nextAt  time.Time
	tmr     timer
}

func New(id string) Scheduler {
	return &scheduler{
		id:        id,
		jobs:      make([]*Job, 0),
		commandCh: make(chan func()),
	}
}

func (s *scheduler) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.running {
		log.Printf("[Scheduler:%s] Starting", s.id)
		s.tmr = newTimer(100 * time.Hour) // arbitrary duration that will be reset in the run loop
		s.running = true
		go s.run()
	}
}

func (s *scheduler) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		log.Printf("[Scheduler:%s] Stopping", s.id)
		s.commandCh <- nil
		s.tmr.Stop()
		s.running = false
	}
}

func (s *scheduler) Running() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.running
}

func (s *scheduler) AddJob(job Job) error {
	if !job.Valid() {
		return ErrInvalidJob
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Printf("[Scheduler:%s] Adding job: %v", s.id, job)
	if s.running {
		s.commandCh <- func() {
			s.addJob(&job)
			s.nextJob = nil // cause the next job to be recalculated
		}
	} else {
		s.addJob(&job)
	}
	return nil
}

func (s *scheduler) RemoveJob(job Job) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		s.commandCh <- func() {
			s.removeJob(&job)
			s.nextJob = nil // cause the next job to be recalculated
		}
	} else {
		s.removeJob(&job)
	}
}

func (s *scheduler) NextJob() *Job {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		retCh := make(chan *Job, 1)
		s.commandCh <- func() {
			retCh <- s.nextJob
		}
		return <-retCh
	}
	_, nextJob := s.next(time_Now().Local())
	return nextJob
}

func (s *scheduler) ReadJobs() []Job {
	result := make([]Job, 0)

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		var wg sync.WaitGroup
		wg.Add(1)
		s.commandCh <- func() {
			for _, j := range s.jobs {
				result = append(result, *j)
			}
			wg.Done()
		}
		wg.Wait()
	} else {
		for _, j := range s.jobs {
			result = append(result, *j)
		}
	}
	return result
}

func (s *scheduler) Override(j Job) {
	s.commandCh <- func() {
		now := time_Now().Local()
		s.nextAt = j.nextOccuranceAfter(now)
		s.nextJob = &j
		s.tmr.Reset(s.nextAt.Sub(now))
		log.Printf("[Scheduler:%s] Override job at %v - %v", s.id, s.nextAt, s.nextJob)
	}
}

func (s *scheduler) CancelOverride() {
	s.commandCh <- func() {
		s.nextJob = nil
	}
}

func (s *scheduler) run() {
	s.setCurrentState()
	for {
		if s.nextJob == nil {
			now := time_Now().Local()
			s.nextAt, s.nextJob = s.next(now)
			s.tmr.Reset(s.nextAt.Sub(now))
			log.Printf("[Scheduler:%s] Next job at %v - %v", s.id, s.nextAt, s.nextJob)
		}
		select {
		case <-s.tmr.Channel():
			if s.nextJob != nil {
				go s.nextJob.Action()
				s.nextJob = nil
			}
		case f := <-s.commandCh:
			if f == nil {
				// Scheduler is stopping. Exit.
				return
			}
			f()
		}
	}
}

func (s *scheduler) addJob(j *Job) {
	s.jobs = append(s.jobs, j)
	sortJobs(s.jobs)
}

func (s *scheduler) removeJob(job *Job) {
	newJobs := make([]*Job, 0)
	for _, j := range s.jobs {
		if j.Hour != job.Hour || j.Min != job.Min || j.Label != job.Label {
			newJobs = append(newJobs, j)
		}
	}
	s.jobs = newJobs
}

func (s *scheduler) setCurrentState() {
	if len(s.jobs) < 1 {
		return
	}
	hour, min, _ := time_Now().Local().Clock()
	var previous *Job
	for _, j := range s.jobs {
		if j.after(hour, min) {
			break
		}
		previous = j
	}
	if previous == nil {
		previous = s.jobs[len(s.jobs)-1]
	}
	go previous.Action()
}

func (s *scheduler) next(now time.Time) (time.Time, *Job) {
	if len(s.jobs) < 1 {
		return now.Add(24 * time.Hour), nil
	}
	hour, min, _ := now.Clock()
	for _, job := range s.jobs {
		if job.after(hour, min) {
			return job.nextOccuranceAfter(now), job
		}
	}
	return s.jobs[0].nextOccuranceAfter(now), s.jobs[0]
}
