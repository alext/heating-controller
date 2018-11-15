package controller

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alext/heating-controller/scheduler"
	"github.com/alext/heating-controller/units"
)

type Event struct {
	Time        units.TimeOfDay   `json:"time"`
	Action      Action            `json:"action"`
	ThermAction *ThermostatAction `json:"therm_action,omitempty"`
}

func (e Event) Valid() bool {
	return e.Time.Valid()
}

func (e Event) NextOccurance() time.Time {
	return e.Time.NextOccuranceAfter(timeNow().Local())
}

func (e Event) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", e.Time, e.Action)
	if e.ThermAction != nil {
		fmt.Fprintf(&b, " %s", e.ThermAction)
	}
	return b.String()
}

func (e Event) buildSchedulerJob(demand func(Event)) scheduler.Job {
	return scheduler.Job{
		Time:   e.Time,
		Label:  e.Action.String(),
		Action: func() { demand(e) },
	}
}

func sortEvents(events []Event) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].Time < events[j].Time
	})
}
