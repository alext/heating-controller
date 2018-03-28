package controller_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/output"
)

var _ = Describe("A heating zone", func() {

	Describe("constructing a zone", func() {
		var (
			out output.Output
		)

		BeforeEach(func() {
			out = output.Virtual("anything")
		})

		It("should return a zone with the given id", func() {
			Expect(controller.NewZone("foo", out).ID).To(Equal("foo"))
		})

		It("should construct a scheduler", func() {
			z := controller.NewZone("foo", out)
			Expect(z.Scheduler).NotTo(BeNil())
		})
	})

	Describe("adding, removing and reading events", func() {
		var (
			z *controller.Zone
		)

		BeforeEach(func() {
			z = controller.NewZone("someting", output.Virtual("something"))
		})

		It("should allow adding and reading events", func() {
			Expect(
				z.AddEvent(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				z.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}),
			).To(Succeed())

			events := z.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events).To(ContainElement(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}))
			Expect(events).To(ContainElement(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}))
		})

		It("should sort the events by time when adding", func() {
			Expect(
				z.AddEvent(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				z.AddEvent(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				z.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}),
			).To(Succeed())

			events := z.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events[0]).To(Equal(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}))
			Expect(events[1]).To(Equal(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}))
			Expect(events[2]).To(Equal(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOn}))
		})

		It("should return an error if an invalid event is added", func() {
			Expect(
				z.AddEvent(controller.Event{Hour: 24, Min: 15, Action: controller.TurnOn}),
			).NotTo(Succeed())
		})

		It("should allow removing an event", func() {
			Expect(
				z.AddEvent(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				z.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}),
			).To(Succeed())
			Expect(
				z.AddEvent(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOff}),
			).To(Succeed())

			z.RemoveEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff})

			events := z.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events).NotTo(ContainElement(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}))
		})

		It("should return a copy of the events list", func() {
			Expect(
				z.AddEvent(controller.Event{Hour: 6, Min: 15, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				z.AddEvent(controller.Event{Hour: 12, Min: 0, Action: controller.TurnOn}),
			).To(Succeed())
			Expect(
				z.AddEvent(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOff}),
			).To(Succeed())

			events := z.ReadEvents()

			// Event will be added at index 2, which would overwrite the above returned slice if it's not a copy
			Expect(
				z.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}),
			).To(Succeed())

			Expect(events).To(HaveLen(3))
			Expect(events[2]).To(Equal(controller.Event{Hour: 18, Min: 0, Action: controller.TurnOff}))
		})
	})
})
