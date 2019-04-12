package controller

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/scheduler/schedulerfakes"
	"github.com/alext/heating-controller/units"
)

var _ = Describe("EventHandler", func() {
	Describe("adding, removing and reading events", func() {
		var (
			sched *schedulerfakes.FakeScheduler
			eh    EventHandler
		)

		BeforeEach(func() {
			sched = &schedulerfakes.FakeScheduler{}
			eh = NewEventHandler(sched, func(Event) {})
		})

		It("should allow adding and reading events", func() {
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(6, 15), Action: On}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}),
			).To(Succeed())

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events).To(ContainElement(Event{Time: units.NewTimeOfDay(6, 15), Action: On}))
			Expect(events).To(ContainElement(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}))
		})

		It("should sort the events by time when adding", func() {
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(6, 15), Action: On}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(18, 0), Action: On}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}),
			).To(Succeed())

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events[0]).To(Equal(Event{Time: units.NewTimeOfDay(6, 15), Action: On}))
			Expect(events[1]).To(Equal(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}))
			Expect(events[2]).To(Equal(Event{Time: units.NewTimeOfDay(18, 0), Action: On}))
		})

		It("should return an error if an invalid event is added", func() {
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(24, 15), Action: On}),
			).NotTo(Succeed())
		})

		Describe("finding an event by time", func() {
			BeforeEach(func() {
				Expect(eh.AddEvent(Event{Time: units.NewTimeOfDay(6, 15), Action: On})).To(Succeed())
				Expect(eh.AddEvent(Event{Time: units.NewTimeOfDay(8, 30), Action: Off})).To(Succeed())
				Expect(eh.AddEvent(Event{Time: units.NewTimeOfDay(18, 0), Action: Off})).To(Succeed())
			})

			It("returns the event with the given time and true", func() {
				e, ok := eh.FindEvent(units.NewTimeOfDay(8, 30))
				Expect(ok).To(BeTrue())
				Expect(e).To(Equal(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}))
			})

			It("returns false if no event matches the given time", func() {
				_, ok := eh.FindEvent(units.NewTimeOfDay(8, 45))
				Expect(ok).To(BeFalse())
			})
		})

		Describe("replacing an event", func() {
			BeforeEach(func() {
				Expect(eh.AddEvent(Event{Time: units.NewTimeOfDay(6, 15), Action: On})).To(Succeed())
				Expect(eh.AddEvent(Event{Time: units.NewTimeOfDay(8, 30), Action: Off})).To(Succeed())
				Expect(eh.AddEvent(Event{Time: units.NewTimeOfDay(18, 0), Action: Off})).To(Succeed())
			})

			It("should allow replacing an event", func() {
				Expect(
					eh.ReplaceEvent(units.NewTimeOfDay(8, 30), Event{Time: units.NewTimeOfDay(8, 45), Action: Off}),
				).To(Succeed())
				events := eh.ReadEvents()
				Expect(events).To(HaveLen(3))
				Expect(events).To(ContainElement(Event{Time: units.NewTimeOfDay(8, 45), Action: Off}))
				Expect(events).NotTo(ContainElement(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}))
			})

			It("should re-sort the events after replacing", func() {
				Expect(
					eh.ReplaceEvent(units.NewTimeOfDay(8, 30), Event{Time: units.NewTimeOfDay(19, 45), Action: Off}),
				).To(Succeed())
				events := eh.ReadEvents()
				Expect(events).To(HaveLen(3))
				Expect(events[2]).To(Equal(Event{Time: units.NewTimeOfDay(19, 45), Action: Off}))
			})

			It("should error if the target event doesn't exist", func() {
				Expect(
					eh.ReplaceEvent(units.NewTimeOfDay(9, 30), Event{Time: units.NewTimeOfDay(8, 45), Action: Off}),
				).NotTo(Succeed())
			})

			It("should send the updated events list to the scheduler", func() {
				Expect(
					eh.ReplaceEvent(units.NewTimeOfDay(8, 30), Event{Time: units.NewTimeOfDay(8, 45), Action: Off}),
				).To(Succeed())
				Expect(sched.SetJobsCallCount()).To(Equal(1))
				jobs := sched.SetJobsArgsForCall(0)
				Expect(jobs).To(HaveLen(3))
				Expect(jobs[0].Time).To(Equal(units.NewTimeOfDay(6, 15)))
				Expect(jobs[1].Time).To(Equal(units.NewTimeOfDay(8, 45)))
				Expect(jobs[2].Time).To(Equal(units.NewTimeOfDay(18, 0)))
			})
		})

		It("should allow removing an event", func() {
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(6, 15), Action: On}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(18, 0), Action: Off}),
			).To(Succeed())

			Expect(eh.RemoveEvent(units.NewTimeOfDay(8, 30))).To(Succeed())

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events).NotTo(ContainElement(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}))
		})

		It("should return a copy of the events list", func() {
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(6, 15), Action: On}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(12, 0), Action: On}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(18, 0), Action: Off}),
			).To(Succeed())

			events := eh.ReadEvents()

			// Event will be added at index 2, which would overwrite the above returned slice if it's not a copy
			Expect(
				eh.AddEvent(Event{Time: units.NewTimeOfDay(8, 30), Action: Off}),
			).To(Succeed())

			Expect(events).To(HaveLen(3))
			Expect(events[2]).To(Equal(Event{Time: units.NewTimeOfDay(18, 0), Action: Off}))
		})
	})

	Describe("querying the next event", func() {
		var (
			sched *schedulerfakes.FakeScheduler
			eh    EventHandler
		)

		BeforeEach(func() {
			sched = &schedulerfakes.FakeScheduler{}
			eh = NewEventHandler(sched, func(Event) {})
		})

		It("returns nil with an empty scheduler", func() {
			sched.NextJobReturns(nil)
			Expect(eh.NextEvent()).To(BeNil())
		})

		Context("with some events", func() {
			var e1, e2 Event

			BeforeEach(func() {
				e1 = Event{Time: units.NewTimeOfDay(6, 30), Action: On, ThermAction: &ThermostatAction{Action: SetTarget, Param: 19000}}
				e2 = Event{Time: units.NewTimeOfDay(8, 15), Action: Off}
				eh.AddEvent(e1)
				eh.AddEvent(e2)
			})

			It("returns the event corresponding to the next scheduler job", func() {
				job := e2.buildSchedulerJob(func(e Event) {})
				sched.NextJobReturns(&job)
				Expect(*eh.NextEvent()).To(Equal(e2))
			})

			It("includes all the event detail", func() {
				job := e1.buildSchedulerJob(func(e Event) {})
				sched.NextJobReturns(&job)
				Expect(*eh.NextEvent()).To(Equal(e1))
			})
			It("when boosted it returns a dummy event representing the end of the boost", func() {
				job := Event{Time: units.NewTimeOfDay(16, 12)}.buildSchedulerJob(func(e Event) {})
				sched.NextJobReturns(&job)
				Expect(*eh.NextEvent()).To(Equal(Event{Time: units.NewTimeOfDay(16, 12)}))
			})
		})
	})

	Describe("bost function", func() {
		var (
			mockNow     time.Time
			sched       *schedulerfakes.FakeScheduler
			eh          EventHandler
			activations []Event
		)

		BeforeEach(func() {
			timeNow = func() time.Time {
				return mockNow
			}
			activations = nil
			sched = new(schedulerfakes.FakeScheduler)
			eh = NewEventHandler(sched, func(e Event) {
				activations = append(activations, e)
			})
			eh.AddEvent(Event{Time: units.NewTimeOfDay(6, 15), Action: On})
			eh.AddEvent(Event{Time: units.NewTimeOfDay(8, 0), Action: Off})
			eh.AddEvent(Event{Time: units.NewTimeOfDay(15, 30), Action: On})
			eh.AddEvent(Event{Time: units.NewTimeOfDay(22, 0), Action: Off})
		})

		Describe("boosting", func() {
			It("activates and schedules a deactivation", func() {
				mockNow = todayAt(9, 0, 0)
				eh.Boost(30 * time.Minute)

				Expect(activations).To(HaveLen(1))
				Expect(activations[0].Action).To(Equal(On))
				Expect(eh.Boosted()).To(BeTrue())

				Expect(sched.OverrideCallCount()).To(Equal(1))
				j := sched.OverrideArgsForCall(0)
				Expect(j.Time).To(Equal(units.NewTimeOfDay(9, 30)))
				j.Action()
				Expect(activations).To(HaveLen(2))
				Expect(activations[1].Action).To(Equal(Off))
				Expect(eh.Boosted()).To(BeFalse())
			})

			It("does not schedule a deactivation if there's already an activation within the duration", func() {
				mockNow = todayAt(15, 0, 0)
				eh.Boost(time.Hour)

				Expect(activations).To(HaveLen(1))
				Expect(activations[0].Action).To(Equal(On))

				Expect(sched.OverrideCallCount()).To(Equal(0))
			})

			It("activates and does not schedule a deactivation if called with 0 duration", func() {
				mockNow = todayAt(12, 0, 0)
				eh.Boost(0)

				Expect(activations).To(HaveLen(1))
				Expect(activations[0].Action).To(Equal(On))

				Expect(sched.OverrideCallCount()).To(Equal(0))
			})
		})

		Describe("cancelling the boost", func() {
			It("cancels the scheduler override", func() {
				mockNow = todayAt(14, 0, 0)
				eh.Boost(30 * time.Minute)
				activations = nil

				mockNow = todayAt(14, 10, 0)
				eh.CancelBoost()

				Expect(sched.CancelOverrideCallCount()).To(Equal(1))
			})

			It("restores the initial zone state", func() {
				mockNow = todayAt(14, 0, 0)
				eh.Boost(30 * time.Minute)
				activations = nil

				mockNow = todayAt(14, 10, 0)
				eh.CancelBoost()

				Expect(activations).To(HaveLen(1))
				Expect(activations[0]).To(Equal(Event{Time: units.NewTimeOfDay(8, 0), Action: Off}))
			})

			It("deactivates the zone if there are no events", func() {
				eh = NewEventHandler(sched, func(e Event) {
					activations = append(activations, e)
				})
				mockNow = todayAt(14, 0, 0)
				eh.Boost(30 * time.Minute)
				activations = nil

				mockNow = todayAt(14, 10, 0)
				eh.CancelBoost()

				Expect(activations).To(HaveLen(1))
				Expect(activations[0].Action).To(Equal(Off))
			})

			It("it does nothing if the zone isn't boosted", func() {

				eh.CancelBoost()
				Expect(sched.CancelOverrideCallCount()).To(Equal(0))
				Expect(activations).To(HaveLen(0))
			})
		})

		Describe("reading the boost state", func() {
			It("is not boosted by default", func() {
				Expect(eh.Boosted()).To(BeFalse())
			})

			It("is boosted once boosted", func() {
				mockNow = todayAt(13, 0, 0)
				eh.Boost(30 * time.Minute)

				Expect(eh.Boosted()).To(BeTrue())
			})

			It("is not boosted once the scheduleld deactivation has triggered", func() {
				mockNow = todayAt(13, 0, 0)
				eh.Boost(30 * time.Minute)

				mockNow = todayAt(13, 30, 0)
				Expect(sched.OverrideCallCount()).To(Equal(1))
				se := sched.OverrideArgsForCall(0)
				se.Action()

				Expect(eh.Boosted()).To(BeFalse())
			})

			It("is not boosted once the next non-override event has fired", func() {
				mockNow = todayAt(15, 0, 0)
				eh.Boost(45 * time.Minute)

				mockNow = todayAt(15, 30, 0)
				Expect(sched.AddJobCallCount()).To(Equal(4)) // From BeforeEach
				nextJob := sched.AddJobArgsForCall(2)
				nextJob.Action()

				Expect(eh.Boosted()).To(BeFalse())
			})

			It("is not boosted when the boost has been cancelled", func() {
				mockNow = todayAt(13, 0, 0)
				eh.Boost(30 * time.Minute)

				mockNow = todayAt(13, 10, 0)
				eh.CancelBoost()

				Expect(eh.Boosted()).To(BeFalse())
			})
		})
	})
})
