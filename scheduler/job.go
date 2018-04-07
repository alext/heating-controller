package scheduler

import (
	"fmt"
	"sort"
	"time"
)

type Job struct {
	Hour   int
	Min    int
	Label  string
	Action func()
}

func (j Job) Valid() bool {
	return j.Hour >= 0 && j.Hour < 24 && j.Min >= 0 && j.Min < 60
}

func (j Job) nextOccuranceAfter(current time.Time) time.Time {
	next := time.Date(current.Year(), current.Month(), current.Day(), j.Hour, j.Min, 0, 0, time.Local)
	if next.Before(current) {
		current = current.AddDate(0, 0, 1)
		next = time.Date(current.Year(), current.Month(), current.Day(), j.Hour, j.Min, 0, 0, time.Local)
	}
	return next
}

func (j Job) after(hour, min int) bool {
	return j.Hour > hour || (j.Hour == hour && j.Min > min)
}

func (j Job) String() string {
	return fmt.Sprintf("%d:%02d %s", j.Hour, j.Min, j.Label)
}

func sortJobs(jobs []*Job) {
	sort.Slice(jobs, func(i, j int) bool {
		a, b := jobs[i], jobs[j]
		return a.Hour < b.Hour || (a.Hour == b.Hour && a.Min < b.Min)
	})
}
