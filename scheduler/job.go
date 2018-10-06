package scheduler

import (
	"fmt"
	"sort"

	"github.com/alext/heating-controller/units"
)

type Job struct {
	Time   units.TimeOfDay
	Label  string
	Action func()
}

func (j Job) Valid() bool {
	// FIXME: Add something here, or remove entirely
	return true
}

func (j Job) after(hour, min int) bool {
	return j.Time > units.NewTimeOfDay(hour, min)
}

func (j Job) String() string {
	return fmt.Sprintf("%s %s", j.Time, j.Label)
}

func sortJobs(jobs []*Job) {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Time < jobs[j].Time
	})
}
