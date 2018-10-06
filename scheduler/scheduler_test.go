package scheduler

import (
	"io/ioutil"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/alext/heating-controller/units"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
)

func TestScheduler(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)

	RunSpecs(t, "Scheduler")
}

var (
	timerCh    chan time.Time
	resetParam time.Duration
	waitNotify chan bool
)

type dummyTimer struct{}

func (tmr dummyTimer) Channel() <-chan time.Time {
	// Notify the channel, but don't block if nothing is listening.
	select {
	case waitNotify <- true:
	default:
	}

	return timerCh
}

func (tmr dummyTimer) Reset(d time.Duration) bool {
	resetParam = d
	return true
}

func (tmr dummyTimer) Stop() bool {
	return true
}

type statefulThing struct {
	mu    sync.Mutex
	state bool
}

func (s *statefulThing) TurnOn()     { s.mu.Lock(); defer s.mu.Unlock(); s.state = true }
func (s *statefulThing) TurnOff()    { s.mu.Lock(); defer s.mu.Unlock(); s.state = false }
func (s *statefulThing) State() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.state }

func (s *statefulThing) ExpectState(st bool) {
	// State should change to st
	EventuallyWithOffset(1, s.State, 100*time.Millisecond, time.Millisecond).Should(Equal(st))
	// and remain there
	ConsistentlyWithOffset(1, s.State, 10*time.Millisecond, time.Millisecond).Should(Equal(st))
}

var _ = Describe("a basic scheduler", func() {
	var (
		mockNow      time.Time
		nowCount     int
		theScheduler Scheduler
		thing        *statefulThing

		HaveNextJobLabelled = func(label string) gomegatypes.GomegaMatcher {
			return WithTransform(func(s Scheduler) string {
				e := s.NextJob()
				if s.Running() {
					<-waitNotify
				}
				return e.Label
			}, Equal(label))
		}
	)

	BeforeEach(func() {
		timerCh = make(chan time.Time, 1)
		waitNotify = make(chan bool, 1)

		newTimer = func(d time.Duration) timer {
			return dummyTimer{}
		}

		thing = &statefulThing{}
		theScheduler = New("something")

		mockNow = time.Now()
		nowCount = 0
		timeNow = func() time.Time {
			nowCount++
			return mockNow
		}
	})

	AfterEach(func() {
		theScheduler.Stop()
	})

	Describe("starting and stopping the scheduler", func() {
		It("should not be running when newly created", func() {
			Expect(theScheduler.Running()).To(BeFalse())
		})

		It("should start the scheduler", func() {
			theScheduler.Start()
			Expect(theScheduler.Running()).To(BeTrue())
		})

		It("should do nothing when attempting to start a running scheduler", func() {
			theScheduler.Start()
			theScheduler.Start()
			<-waitNotify

			Expect(nowCount).To(Equal(1))
		})

		It("should stop the scheduler", func() {
			theScheduler.Start()
			theScheduler.Stop()
			Expect(theScheduler.Running()).To(BeFalse())
		})

		It("should do nothing when attempting to stop a non-running scheduler", func(done Done) {
			theScheduler.Stop()
			close(done)
		}, 0.5)

		Describe("setting the initial output state", func() {
			Context("with some entries", func() {
				BeforeEach(func() {
					theScheduler.AddJob(Job{Time: units.NewTimeOfDay(6, 30), Action: thing.TurnOn})
					theScheduler.AddJob(Job{Time: units.NewTimeOfDay(7, 45), Action: thing.TurnOff})
					theScheduler.AddJob(Job{Time: units.NewTimeOfDay(17, 33), Action: thing.TurnOn})
					theScheduler.AddJob(Job{Time: units.NewTimeOfDay(21, 12), Action: thing.TurnOff})
				})

				It("should apply the previous entry's state on starting", func() {
					mockNow = todayAt(6, 45, 0)

					theScheduler.Start()
					<-waitNotify
					thing.ExpectState(true)
					theScheduler.Stop()

					mockNow = todayAt(12, 00, 0)
					theScheduler.Start()
					<-waitNotify
					thing.ExpectState(false)
				})

				It("should use the last entry from the previous day if necessary", func() {
					mockNow = todayAt(4, 45, 0)

					theScheduler.Start()
					<-waitNotify
					thing.ExpectState(false)
				})
			})

			It("should do nothing with no entries", func() {
				theScheduler.Start()
				<-waitNotify
				// expect it not to blow up
			})
		})
	})

	It("should continuously sleep for a day when started with no entries", func() {
		mockNow = todayAt(6, 20, 0)

		theScheduler.Start()
		<-waitNotify

		Expect(resetParam.String()).To(Equal("24h0m0s"))

		mockNow = mockNow.AddDate(0, 0, 1)
		timerCh <- mockNow
		<-waitNotify

		Expect(resetParam.String()).To(Equal("24h0m0s"))
	})

	Describe("firing jobs as scheduled", func() {

		BeforeEach(func() {
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(6, 30), Action: thing.TurnOn})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(7, 45), Action: thing.TurnOff})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(17, 33), Action: thing.TurnOn})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(21, 12), Action: thing.TurnOff})
		})

		It("should fire the given jobs in order", func() {
			mockNow = todayAt(6, 20, 0)

			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("10m0s"))

			mockNow = todayAt(6, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("1h15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("9h48m0s"))

			mockNow = todayAt(17, 33, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)
		})

		It("should wrap around at the end of the day", func() {
			mockNow = todayAt(20, 04, 23)

			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("1h7m37s"))

			mockNow = todayAt(21, 12, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			nextAt := todayAt(6, 30, 0).AddDate(0, 0, 1)

			Expect(resetParam.String()).To(Equal(nextAt.Sub(mockNow).String()))

			mockNow = nextAt
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)
		})

		It("should handle jobs added in a non-sequential order", func() {
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(13, 00), Action: thing.TurnOff})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(11, 30), Action: thing.TurnOn})

			mockNow = todayAt(7, 30, 0)

			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("3h45m0s"))

			mockNow = todayAt(11, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("1h30m0s"))

			mockNow = todayAt(13, 0, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("4h33m0s"))
		})

		It("should handle jobs added after the scheduler has been started", func() {
			mockNow = todayAt(7, 30, 0)

			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("9h48m0s"))

			mockNow = todayAt(9, 30, 0)
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(11, 30), Action: thing.TurnOn})
			<-waitNotify

			Expect(resetParam.String()).To(Equal("2h0m0s"))

			mockNow = todayAt(11, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)
		})
	})

	It("should return an error when adding an invalid job", func() {
		err := theScheduler.AddJob(Job{Time: units.NewTimeOfDay(25, 0)})
		Expect(err).To(MatchError(ErrInvalidJob))
		Expect(theScheduler.ReadJobs()).To(HaveLen(0))
	})

	Describe("querying the next job", func() {

		It("should return nil with no jobs", func() {
			Expect(theScheduler.NextJob()).To(BeNil())
		})

		Context("with some jobs", func() {
			BeforeEach(func() {
				theScheduler.AddJob(Job{Time: units.NewTimeOfDay(6, 30), Action: thing.TurnOn, Label: "alpha"})
				theScheduler.AddJob(Job{Time: units.NewTimeOfDay(17, 33), Action: thing.TurnOn, Label: "bravo"})
				theScheduler.AddJob(Job{Time: units.NewTimeOfDay(7, 45), Action: thing.TurnOff, Label: "charlie"})
				theScheduler.AddJob(Job{Time: units.NewTimeOfDay(21, 12), Action: thing.TurnOff, Label: "delta"})
			})

			It("should return the next job", func() {
				mockNow = todayAt(6, 0, 0)

				Expect(theScheduler).To(HaveNextJobLabelled("alpha"))

				mockNow = todayAt(7, 30, 0)

				Expect(theScheduler).To(HaveNextJobLabelled("charlie"))
			})

			It("should handle the wrap around at the end of the day", func() {
				mockNow = todayAt(21, 30, 0)

				Expect(theScheduler).To(HaveNextJobLabelled("alpha"))
			})

			Context("with a running timer", func() {
				BeforeEach(func() {
					mockNow = todayAt(14, 0, 0)
					theScheduler.Start()
					<-waitNotify
				})

				It("should return the next job", func() {
					Expect(theScheduler).To(HaveNextJobLabelled("bravo"))
				})

				It("should return the override job when boosted", func() {
					theScheduler.Override(Job{Time: units.NewTimeOfDay(14, 30), Action: thing.TurnOn})
					<-waitNotify

					e := theScheduler.NextJob()
					Expect(e.Time).To(Equal(units.NewTimeOfDay(14, 30)))
				})
			})
		})
	})

	Describe("reading the current schedule", func() {

		It("should return an empty list for a stopped scheduler with no jobs", func() {
			Expect(theScheduler.ReadJobs()).To(BeEmpty())
		})

		It("should return an empty list for a running scheduler with no jobs", func() {
			theScheduler.Start()
			<-waitNotify

			Expect(theScheduler.ReadJobs()).To(BeEmpty())
		})

		Context("with some jobs", func() {
			BeforeEach(func() {
				theScheduler.AddJob(Job{Time: units.NewTimeOfDay(6, 30), Action: thing.TurnOn, Label: "alpha"})
				theScheduler.AddJob(Job{Time: units.NewTimeOfDay(17, 33), Action: thing.TurnOn, Label: "bravo"})
				theScheduler.AddJob(Job{Time: units.NewTimeOfDay(7, 45), Action: thing.TurnOff, Label: "charlie"})
				theScheduler.AddJob(Job{Time: units.NewTimeOfDay(21, 12), Action: thing.TurnOff, Label: "delta"})
			})

			It("should return the current job list", func() {
				jobs := theScheduler.ReadJobs()

				Expect(jobs).To(HaveLen(4))
				Expect(jobs[0].Label).To(Equal("alpha"))
				Expect(jobs[3].Label).To(Equal("delta"))
			})

			It("should return the current job list from a running scheduler", func() {
				theScheduler.Start()
				<-waitNotify

				jobs := theScheduler.ReadJobs()

				Expect(jobs).To(HaveLen(4))
				Expect(jobs[0].Label).To(Equal("alpha"))
				Expect(jobs[3].Label).To(Equal("delta"))
			})
		})

	})

	Describe("removing a job", func() {
		BeforeEach(func() {
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(6, 30), Action: thing.TurnOn, Label: "alpha"})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(17, 33), Action: thing.TurnOn, Label: "bravo"})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(7, 45), Action: thing.TurnOff, Label: "charlie"})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(21, 12), Action: thing.TurnOff, Label: "delta"})
		})

		Context("with a stopped timer", func() {
			It("should remove the corresponding job from the list", func() {
				theScheduler.RemoveJob(Job{Time: units.NewTimeOfDay(7, 45), Action: thing.TurnOff, Label: "charlie"})

				Expect(theScheduler.ReadJobs()).To(HaveLen(3))
			})

			It("should do nothing if the job isn't in the scheduler", func() {
				theScheduler.RemoveJob(Job{Time: units.NewTimeOfDay(7, 50), Action: thing.TurnOn, Label: "foo"})

				Expect(theScheduler.ReadJobs()).To(HaveLen(4))
			})
		})

		Context("with a running scheduler", func() {
			BeforeEach(func() {
				mockNow = todayAt(14, 0, 0)
				theScheduler.Start()
				<-waitNotify
			})

			It("should remove the job from the list", func() {
				theScheduler.RemoveJob(Job{Time: units.NewTimeOfDay(7, 45), Action: thing.TurnOff, Label: "charlie"})

				Expect(theScheduler.ReadJobs()).To(HaveLen(3))
			})

			It("should do nothing if the job isn't in the scheduler", func() {
				theScheduler.RemoveJob(Job{Time: units.NewTimeOfDay(7, 50), Action: thing.TurnOn, Label: "foo"})

				Expect(theScheduler.ReadJobs()).To(HaveLen(4))
			})

			It("should reschedule if the removed job was the next job", func() {
				mockNow = todayAt(15, 0, 0)
				theScheduler.RemoveJob(Job{Time: units.NewTimeOfDay(17, 33), Action: thing.TurnOn, Label: "bravo"})

				<-waitNotify
				Expect(resetParam.String()).To(Equal("6h12m0s"))
			})
		})
	})

	Describe("override function", func() {
		BeforeEach(func() {
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(6, 30), Action: thing.TurnOn, Label: "alpha"})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(7, 45), Action: thing.TurnOff, Label: "bravo"})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(17, 33), Action: thing.TurnOn, Label: "charlie"})
			theScheduler.AddJob(Job{Time: units.NewTimeOfDay(21, 12), Action: thing.TurnOff, Label: "delta"})

			mockNow = todayAt(14, 0, 0)
			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(false)
			Expect(theScheduler).To(HaveNextJobLabelled("charlie"))
		})

		It("schedules the given job as the next one", func() {
			mockNow = todayAt(15, 30, 0)
			theScheduler.Override(Job{Time: units.NewTimeOfDay(16, 30), Action: thing.TurnOn, Label: "override"})

			<-waitNotify
			thing.ExpectState(false)
			Expect(resetParam.String()).To(Equal("1h0m0s"))
			Expect(theScheduler).To(HaveNextJobLabelled("override"))

			mockNow = todayAt(16, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)
			Expect(resetParam.String()).To(Equal("1h3m0s"))
			Expect(theScheduler).To(HaveNextJobLabelled("charlie"))

			Expect(theScheduler.ReadJobs()).To(HaveLen(4))
		})

		It("skips any existing jobs in between", func() {
			mockNow = todayAt(17, 0, 0)
			theScheduler.Override(Job{Time: units.NewTimeOfDay(17, 45), Action: thing.TurnOn, Label: "override"})

			<-waitNotify
			thing.ExpectState(false)
			Expect(resetParam.String()).To(Equal("45m0s"))
			Expect(theScheduler).To(HaveNextJobLabelled("override"))

			mockNow = todayAt(17, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)
			Expect(resetParam.String()).To(Equal("3h27m0s"))
			Expect(theScheduler).To(HaveNextJobLabelled("delta"))

			Expect(theScheduler.ReadJobs()).To(HaveLen(4))
		})

		Describe("cancelling the override", func() {
			BeforeEach(func() {
				mockNow = todayAt(16, 0, 0)
				theScheduler.Override(Job{Time: units.NewTimeOfDay(18, 30), Action: thing.TurnOn, Label: "override"})

				<-waitNotify
				thing.ExpectState(false)
				Expect(resetParam.String()).To(Equal("2h30m0s"))
				Expect(theScheduler).To(HaveNextJobLabelled("override"))
			})

			It("cancelling the override continues with the next scheduled job", func() {
				mockNow = todayAt(16, 30, 0)
				theScheduler.CancelOverride()

				<-waitNotify
				thing.ExpectState(false)
				Expect(resetParam.String()).To(Equal("1h3m0s"))
				Expect(theScheduler).To(HaveNextJobLabelled("charlie"))
			})

			It("ignores any jobs that have been skipped in the meantime", func() {
				mockNow = todayAt(17, 45, 0)
				theScheduler.CancelOverride()

				<-waitNotify
				thing.ExpectState(false)
				Expect(resetParam.String()).To(Equal("3h27m0s"))
				Expect(theScheduler).To(HaveNextJobLabelled("delta"))
			})
		})
	})
})

func todayAt(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, time.Local)
}
