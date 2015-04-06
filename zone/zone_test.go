package zone

import (
	"io/ioutil"
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/scheduler"
)

func TestZone(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)

	RunSpecs(t, "Zone")
}

var _ = Describe("A heating zone", func() {

	Describe("constructing a zone", func() {
		var (
			out output.Output
		)

		BeforeEach(func() {
			out = output.Virtual("anything")
		})

		It("should return a zone with the given id", func() {
			Expect(New("foo", out).ID).To(Equal("foo"))
		})

		It("should construct a scheduler", func() {
			z := New("foo", out)
			Expect(z.Scheduler).NotTo(BeNil())
		})

		It("should set the override state to Normal by default", func() {
			Expect(New("foo", out).overrideMode).To(Equal(ModeNormal))
		})
	})

	Describe("setting zone override state", func() {
		var (
			out output.Output
			z   *Zone
		)

		BeforeEach(func() {
			out = output.Virtual("something")
			z = New("someting", out)
		})

		It("should activate the output when set to override on", func() {
			z.SetOverride(ModeOverrideOn)

			Expect(out.Active()).To(BeTrue())
		})

		It("should deactivate the output when set to override off", func() {
			out.Activate()

			z.SetOverride(ModeOverrideOff)

			Expect(out.Active()).To(BeFalse())
		})

		It("should respect the demand state when set to normal mode", func() {
			z.SetOverride(ModeOverrideOff)
			z.schedulerDemand(scheduler.TurnOn)

			Expect(out.Active()).To(BeFalse())

			z.SetOverride(ModeNormal)

			Expect(out.Active()).To(BeTrue())
		})
	})

	Describe("handling scheduler demand", func() {
		var (
			out output.Output
			z   *Zone
		)

		BeforeEach(func() {
			out = output.Virtual("something")
			z = New("someting", out)
		})

		It("should activate the output when demand is activated", func() {
			z.schedulerDemand(scheduler.TurnOn)

			Expect(out.Active()).To(BeTrue())
		})

		It("should deactivate the output when demand is deactivated", func() {
			out.Deactivate()

			z.schedulerDemand(scheduler.TurnOff)

			Expect(out.Active()).To(BeFalse())
		})

		Context("when the zone is set to override On", func() {
			BeforeEach(func() {
				z.SetOverride(ModeOverrideOn)
			})

			It("should leave the output on when demand is activated", func() {
				z.schedulerDemand(scheduler.TurnOn)

				Expect(out.Active()).To(BeTrue())
			})

			It("should leave the output on when demand is deactivated", func() {
				z.schedulerDemand(scheduler.TurnOff)

				Expect(out.Active()).To(BeTrue())
			})
		})

		Context("when the zone is set to override Off", func() {
			BeforeEach(func() {
				z.SetOverride(ModeOverrideOff)
			})

			It("should leave the output off when demand is activated", func() {
				z.schedulerDemand(scheduler.TurnOn)

				Expect(out.Active()).To(BeFalse())
			})

			It("should leave the output off when demand is deactivated", func() {
				z.schedulerDemand(scheduler.TurnOff)

				Expect(out.Active()).To(BeFalse())
			})
		})
	})
})
