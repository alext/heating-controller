package controller

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/scheduler/schedulerfakes"
)

var _ = Describe("EventHandler", func() {
	Describe("adding, removing and reading events", func() {
		var (
			eh EventHandler
		)

		BeforeEach(func() {
			eh = NewEventHandler(&schedulerfakes.FakeScheduler{}, func(Event) {})
		})

		It("should allow adding and reading events", func() {
			Expect(
				eh.AddEvent(Event{Hour: 6, Min: 15, Action: TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Hour: 8, Min: 30, Action: TurnOff}),
			).To(Succeed())

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events).To(ContainElement(Event{Hour: 6, Min: 15, Action: TurnOn}))
			Expect(events).To(ContainElement(Event{Hour: 8, Min: 30, Action: TurnOff}))
		})

		It("should sort the events by time when adding", func() {
			Expect(
				eh.AddEvent(Event{Hour: 6, Min: 15, Action: TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Hour: 18, Min: 0, Action: TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Hour: 8, Min: 30, Action: TurnOff}),
			).To(Succeed())

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events[0]).To(Equal(Event{Hour: 6, Min: 15, Action: TurnOn}))
			Expect(events[1]).To(Equal(Event{Hour: 8, Min: 30, Action: TurnOff}))
			Expect(events[2]).To(Equal(Event{Hour: 18, Min: 0, Action: TurnOn}))
		})

		It("should return an error if an invalid event is added", func() {
			Expect(
				eh.AddEvent(Event{Hour: 24, Min: 15, Action: TurnOn}),
			).NotTo(Succeed())
		})

		It("should allow removing an event", func() {
			Expect(
				eh.AddEvent(Event{Hour: 6, Min: 15, Action: TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Hour: 8, Min: 30, Action: TurnOff}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Hour: 18, Min: 0, Action: TurnOff}),
			).To(Succeed())

			eh.RemoveEvent(Event{Hour: 8, Min: 30, Action: TurnOff})

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events).NotTo(ContainElement(Event{Hour: 8, Min: 30, Action: TurnOff}))
		})

		It("should return a copy of the events list", func() {
			Expect(
				eh.AddEvent(Event{Hour: 6, Min: 15, Action: TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Hour: 12, Min: 0, Action: TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(Event{Hour: 18, Min: 0, Action: TurnOff}),
			).To(Succeed())

			events := eh.ReadEvents()

			// Event will be added at index 2, which would overwrite the above returned slice if it's not a copy
			Expect(
				eh.AddEvent(Event{Hour: 8, Min: 30, Action: TurnOff}),
			).To(Succeed())

			Expect(events).To(HaveLen(3))
			Expect(events[2]).To(Equal(Event{Hour: 18, Min: 0, Action: TurnOff}))
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
			eh.AddEvent(Event{Hour: 6, Min: 15, Action: TurnOn})
			eh.AddEvent(Event{Hour: 8, Min: 0, Action: TurnOff})
			eh.AddEvent(Event{Hour: 15, Min: 30, Action: TurnOn})
			eh.AddEvent(Event{Hour: 22, Min: 0, Action: TurnOff})
		})

		Describe("boosting", func() {
			It("activates and schedules a deactivation", func() {
				mockNow = todayAt(9, 0, 0)
				eh.Boost(30 * time.Minute)

				Expect(activations).To(HaveLen(1))
				Expect(activations[0].Action).To(Equal(TurnOn))

				Expect(sched.OverrideCallCount()).To(Equal(1))
				j := sched.OverrideArgsForCall(0)
				Expect(j.Hour).To(Equal(9))
				Expect(j.Min).To(Equal(30))
				j.Action()
				Expect(activations).To(HaveLen(2))
				Expect(activations[1].Action).To(Equal(TurnOff))
			})

			It("does not schedule a deactivation if there's already an activation within the duration", func() {
				mockNow = todayAt(15, 0, 0)
				eh.Boost(time.Hour)

				Expect(activations).To(HaveLen(1))
				Expect(activations[0].Action).To(Equal(TurnOn))

				Expect(sched.OverrideCallCount()).To(Equal(0))
			})

			It("activates and does not schedule a deactivation if called with 0 duration", func() {
				mockNow = todayAt(12, 0, 0)
				eh.Boost(0)

				Expect(activations).To(HaveLen(1))
				Expect(activations[0].Action).To(Equal(TurnOn))

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
				Expect(activations[0]).To(Equal(Event{Hour: 8, Min: 0, Action: TurnOff}))
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
				Expect(activations[0].Action).To(Equal(TurnOff))
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
