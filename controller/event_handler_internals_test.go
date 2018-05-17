package controller

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/scheduler/schedulerfakes"
)

var _ = Describe("EventHandler internals", func() {
	var (
		eh      *eventHandler
		mockNow time.Time
	)

	BeforeEach(func() {
		timeNow = func() time.Time { return mockNow }
		eh = &eventHandler{
			sched:  new(schedulerfakes.FakeScheduler),
			demand: func(Event) {},
		}
	})

	Describe("nextEvent", func() {
		It("returns nil with no events", func() {
			Expect(eh.nextEvent()).To(BeNil())
		})

		Context("with some events", func() {
			BeforeEach(func() {
				eh.AddEvent(Event{Hour: 6, Min: 15, Action: On})
				eh.AddEvent(Event{Hour: 8, Min: 30, Action: Off})
				eh.AddEvent(Event{Hour: 18, Min: 0, Action: On})
				eh.AddEvent(Event{Hour: 22, Min: 0, Action: Off})
			})

			It("returns the next event after now", func() {
				mockNow = todayAt(7, 15, 0)
				Expect(eh.nextEvent()).To(Equal(&Event{Hour: 8, Min: 30, Action: Off}))
				mockNow = todayAt(10, 30, 0)
				Expect(eh.nextEvent()).To(Equal(&Event{Hour: 18, Min: 0, Action: On}))
			})

			It("returns the first event if there are no more events today", func() {
				mockNow = todayAt(22, 5, 0)
				Expect(eh.nextEvent()).To(Equal(&Event{Hour: 6, Min: 15, Action: On}))
			})
		})
	})

	Describe("previousEvent", func() {
		It("returns nil with no events", func() {
			Expect(eh.previousEvent()).To(BeNil())
		})

		Context("with some events", func() {
			BeforeEach(func() {
				eh.AddEvent(Event{Hour: 6, Min: 15, Action: On})
				eh.AddEvent(Event{Hour: 8, Min: 30, Action: Off})
				eh.AddEvent(Event{Hour: 18, Min: 0, Action: On})
				eh.AddEvent(Event{Hour: 22, Min: 0, Action: Off})
			})

			It("returns the event before now", func() {
				mockNow = todayAt(7, 15, 0)
				Expect(eh.previousEvent()).To(Equal(&Event{Hour: 6, Min: 15, Action: On}))
				mockNow = todayAt(10, 30, 0)
				Expect(eh.previousEvent()).To(Equal(&Event{Hour: 8, Min: 30, Action: Off}))
			})

			It("returns the last event if there are no earlier events today", func() {
				mockNow = todayAt(6, 5, 0)
				Expect(eh.previousEvent()).To(Equal(&Event{Hour: 22, Min: 0, Action: Off}))
			})
		})
	})
})
