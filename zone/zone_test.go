package zone

import (
	"io/ioutil"
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

		Describe("activating demand", func() {
			BeforeEach(func() {
				out.Deactivate()
				z.schedDemand = false

				z.schedulerDemand(scheduler.TurnOn)
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

				z.schedulerDemand(scheduler.TurnOff)
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
			z = New("someting", out)
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
				Out:           out,
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
