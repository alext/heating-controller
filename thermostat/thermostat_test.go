package thermostat

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/units"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestThermostat(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)

	RunSpecs(t, "Thermostat")
}

var _ = Describe("A Thermostat", func() {
	var (
		t    *thermostat
		sens sensor.SettableSensor
	)

	BeforeEach(func() {
		t = nil
		sens = sensor.NewPushSensor("something")
		sens.Set(19000, time.Now())
	})

	AfterEach(func() {
		if t != nil {
			t.Close()
		}
	})

	Describe("constructing a thermostat", func() {
		It("builds one correctly", func() {
			t = New("something", sens, 19000, func(b bool) {}).(*thermostat)
			Expect(t.id).To(Equal("something"))
			Expect(t.target).To(BeNumerically("==", 19000))
			Expect(t.current).To(BeNumerically("==", 19000))
		})

		It("starts as active when current temp is within the threshold", func() {
			t = New("something", sens, 19000, func(b bool) {}).(*thermostat)
			Expect(t.active).To(BeTrue())
		})
	})

	Describe("setting the target temperature", func() {
		BeforeEach(func() {
			t = &thermostat{
				current: 18000,
				target:  17000,
			}
		})

		It("Updates the target for the thermostat", func() {
			t.Set(16000)
			Expect(t.target).To(BeNumerically("==", 16000))
		})

		It("triggers the thermostat to update the demand state", func() {
			t.Set(19000)
			Expect(t.active).To(BeTrue())
		})
	})

	Describe("subscribing to sensor updates", func() {
		BeforeEach(func() {
			t = New("sonething", sens, 18000, func(b bool) {}).(*thermostat)
		})

		It("should update the current value", func() {
			Expect(t.Current()).To(BeEquivalentTo(19000))

			sens.Set(21000, time.Now())
			Eventually(t.Current).Should(BeEquivalentTo(21000))

			sens.Set(1234, time.Now())
			Eventually(t.Current).Should(BeEquivalentTo(1234))
		})

		It("triggers an update of the current state", func() {
			Expect(t.active).To(BeFalse())

			sens.Set(17000, time.Now())
			Eventually(t.Current).Should(BeEquivalentTo(17000)) // Wait for async update

			Expect(t.active).To(BeTrue())
		})
	})

	type TriggeringCase struct {
		Current            units.Temperature
		Target             units.Temperature
		CurrentlyActive    bool
		ExpectedActive     bool
		ExpectDemandCalled bool
	}

	DescribeTable("triggering changes in state",
		func(c TriggeringCase) {
			demandCalled := false
			demandNotify := make(chan struct{})
			t := &thermostat{
				current: c.Current,
				target:  c.Target,
				active:  c.CurrentlyActive,
				demand: func(param bool) {
					defer GinkgoRecover()
					demandCalled = true
					Expect(param).To(Equal(c.ExpectedActive))
					close(demandNotify)
				},
			}

			t.trigger()
			Expect(t.active).To(Equal(c.ExpectedActive))
			if c.ExpectDemandCalled {
				select {
				case <-demandNotify:
				case <-time.After(time.Second):
				}
				Expect(demandCalled).To(BeTrue(), "expected demandFunc to be called")
			}
		},
		Entry("activates when current well below target", TriggeringCase{
			Current: 15000, Target: 18000, CurrentlyActive: false,
			ExpectedActive: true, ExpectDemandCalled: true,
		}),
		Entry("remains active when current well below target", TriggeringCase{
			Current: 15000, Target: 18000, CurrentlyActive: true,
			ExpectedActive: true, ExpectDemandCalled: false,
		}),
		Entry("deactivates when current well above target", TriggeringCase{
			Current: 20000, Target: 18000, CurrentlyActive: true,
			ExpectedActive: false, ExpectDemandCalled: true,
		}),
		Entry("deactivates when current slightly above target", TriggeringCase{
			Current: 18050, Target: 18000, CurrentlyActive: true,
			ExpectedActive: false, ExpectDemandCalled: true,
		}),
		Entry("remains inactive when current well above target", TriggeringCase{
			Current: 20000, Target: 18000, CurrentlyActive: false,
			ExpectedActive: false, ExpectDemandCalled: false,
		}),
		Entry("remains active when current within threhold below target", TriggeringCase{
			Current: 17950, Target: 18000, CurrentlyActive: true,
			ExpectedActive: true, ExpectDemandCalled: false,
		}),
		Entry("remains inactive when current within threhold below target", TriggeringCase{
			Current: 17950, Target: 18000, CurrentlyActive: false,
			ExpectedActive: false, ExpectDemandCalled: false,
		}),
		Entry("remains inactive when current slightly above target", TriggeringCase{
			Current: 18050, Target: 18000, CurrentlyActive: false,
			ExpectedActive: false, ExpectDemandCalled: false,
		}),
	)
})
