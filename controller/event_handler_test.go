package controller_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/scheduler/schedulerfakes"
)

var _ = Describe("EventHandler", func() {
	Describe("adding, removing and reading events", func() {
		var (
			eh controller.EventHandler
		)

		BeforeEach(func() {
			eh = controller.NewEventHandler(&schedulerfakes.FakeScheduler{}, func(controller.Event) {})
		})

		It("should allow adding and reading events", func() {
			Expect(
				eh.AddEvent(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}),
			).To(Succeed())

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events).To(ContainElement(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}))
			Expect(events).To(ContainElement(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}))
		})

		It("should sort the events by time when adding", func() {
			Expect(
				eh.AddEvent(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}),
			).To(Succeed())

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events[0]).To(Equal(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}))
			Expect(events[1]).To(Equal(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}))
			Expect(events[2]).To(Equal(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOn}))
		})

		It("should return an error if an invalid event is added", func() {
			Expect(
				eh.AddEvent(controller.Event{Hour: 24, Min: 15, Action: controller.TurnOn}),
			).NotTo(Succeed())
		})

		It("should allow removing an event", func() {
			Expect(
				eh.AddEvent(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}),
			).To(Succeed())
			Expect(
				eh.AddEvent(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOff}),
			).To(Succeed())

			eh.RemoveEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff})

			events := eh.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events).NotTo(ContainElement(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}))
		})

		It("should return a copy of the events list", func() {
			Expect(
				eh.AddEvent(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(controller.Event{Hour: 12, Min: 0, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				eh.AddEvent(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOff}),
			).To(Succeed())

			events := eh.ReadEvents()

			// Event will be added at index 2, which would overwrite the above returned slice if it's not a copy
			Expect(
				eh.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}),
			).To(Succeed())

			Expect(events).To(HaveLen(3))
			Expect(events[2]).To(Equal(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOff}))
		})
	})
})
