package controller

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/thermostat/thermostatfakes"
)

var _ = Describe("Zone demand handling", func() {

	Describe("applying Event actions", func() {
		var (
			therm *thermostatfakes.FakeThermostat
			z     *Zone
		)

		BeforeEach(func() {
			therm = &thermostatfakes.FakeThermostat{}
			z = NewZone("something", output.Virtual("something"))
			z.Thermostat = therm
		})

		Context("with no ThermAction", func() {
			It("triggers the scheduler demand", func() {
				z.applyEvent(Event{Action: On})
				Expect(z.schedDemand).To(BeTrue())
				// more detailed tests below
			})
		})

		Context("with a ThermAction", func() {
			var e Event
			BeforeEach(func() { e = Event{Action: On, ThermAction: &ThermostatAction{Action: SetTarget, Param: 19000}} })

			It("still triggers the scheduler demand", func() {
				z.applyEvent(e)
				Expect(z.schedDemand).To(BeTrue())
			})

			It("sets the thermostat target", func() {
				z.applyEvent(e)
				Expect(therm.SetCallCount()).To(Equal(1))
				// more detailed tests in action_test.go
			})

			It("ignores the ThermAction if the zone has no thermostat", func() {
				z.Thermostat = nil
				z.applyEvent(e)

				//should still trigger the scheduler
				Expect(z.schedDemand).To(BeTrue())

				//should not blow up...
			})
		})
	})

	Describe("handling scheduler demand", func() {
		var (
			out output.Output
			z   *Zone
		)

		BeforeEach(func() {
			out = output.Virtual("something")
			z = NewZone("someting", out)
		})

		Describe("activating demand", func() {
			BeforeEach(func() {
				out.Deactivate()
				z.schedDemand = false

				z.schedulerDemand(true)
			})

			It("should update the demand state", func() {
				Expect(z.schedDemand).To(BeTrue())
			})

			It("should trigger an update to overall demand", func() {
				Expect(out.Active()).To(BeTrue())
			})
		})

		Describe("deactivating demand", func() {
			BeforeEach(func() {
				out.Activate()
				z.schedDemand = true
				z.currentDemand = true // force the output to be updated

				z.schedulerDemand(false)
			})

			It("should update the demand state", func() {
				Expect(z.schedDemand).To(BeFalse())
			})

			It("should trigger an update to overall demand", func() {
				Expect(out.Active()).To(BeFalse())
			})
		})
	})

	Describe("handling thermostat demand", func() {
		var (
			out output.Output
			z   *Zone
		)

		BeforeEach(func() {
			out = output.Virtual("something")
			z = NewZone("someting", out)
			z.schedDemand = true // so that thermDemand changes trigger output
		})

		Describe("activating demand", func() {
			BeforeEach(func() {
				out.Deactivate()
				z.thermDemand = false

				z.thermostatDemand(true)
			})

			It("should update the demand state", func() {
				Expect(z.thermDemand).To(BeTrue())
			})

			It("should trigger an update to overall demand", func() {
				Expect(out.Active()).To(BeTrue())
			})
		})

		Describe("deactivating demand", func() {
			BeforeEach(func() {
				out.Activate()
				z.thermDemand = true
				z.currentDemand = true // force the output to be updated

				z.thermostatDemand(false)
			})

			It("should update the demand state", func() {
				Expect(z.thermDemand).To(BeFalse())
			})

			It("should trigger an update to overall demand", func() {
				Expect(out.Active()).To(BeFalse())
			})
		})
	})

	type demandCase struct {
		schedDemand    bool
		thermDemand    bool
		currentDemand  bool
		expectedDemand bool

		initialOutputState  bool
		expectedOutputState bool
	}

	DescribeTable("applying changes in demand state",
		func(c demandCase) {
			out := output.Virtual("something")
			if c.initialOutputState {
				out.Activate()
			} else {
				out.Deactivate()
			}
			z := &Zone{
				out:           out,
				schedDemand:   c.schedDemand,
				thermDemand:   c.thermDemand,
				currentDemand: c.currentDemand,
			}
			z.updateDemand()
			Expect(z.currentDemand).To(Equal(c.expectedDemand))
			Expect(out.Active()).To(Equal(c.expectedOutputState))
		},
		Entry("activates when both demands active", demandCase{
			schedDemand: true, thermDemand: true, currentDemand: false,
			expectedDemand:     true,
			initialOutputState: false, expectedOutputState: true,
		}),
		Entry("deactivates when only scheduler demand active", demandCase{
			schedDemand: true, thermDemand: false, currentDemand: true,
			expectedDemand:     false,
			initialOutputState: true, expectedOutputState: false,
		}),
		Entry("deactivates when only thermostat demand active", demandCase{
			schedDemand: false, thermDemand: true, currentDemand: true,
			expectedDemand:     false,
			initialOutputState: true, expectedOutputState: false,
		}),
		Entry("deactivates when neither demand active", demandCase{
			schedDemand: false, thermDemand: false, currentDemand: true,
			expectedDemand:     false,
			initialOutputState: true, expectedOutputState: false,
		}),
		Entry("does not change output when demand already active", demandCase{
			schedDemand: true, thermDemand: true, currentDemand: true,
			expectedDemand: true,
			// Set to false to ensure output doesn't get called.
			initialOutputState: false, expectedOutputState: false,
		}),
		Entry("does not change output when demand already inactive", demandCase{
			schedDemand: false, thermDemand: false, currentDemand: false,
			expectedDemand: false,
			// Set to true to ensure output doesn't get called.
			initialOutputState: true, expectedOutputState: true,
		}),
	)
})
