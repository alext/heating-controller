package controller

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/config"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/sensor"
)

var _ = Describe("Controller", func() {

	Describe("Setting up zones", func() {
		var (
			ctrl *Controller
			cfg  map[string]config.ZoneConfig
		)

		BeforeEach(func() {
			DataDir, _ = ioutil.TempDir("", "heating-controller-test")

			cfg = make(map[string]config.ZoneConfig)
			ctrl = New()

			outputNew = func(id string, pin int) (output.Output, error) {
				out := output.Virtual(fmt.Sprintf("%s-gpio%d", id, pin))
				return out, nil
			}
		})
		AfterEach(func() {
			for _, z := range ctrl.Zones {
				z.Scheduler.Stop()
			}
			os.RemoveAll(DataDir)
		})

		It("Should do nothing with a blank list of zones", func() {
			Expect(ctrl.SetupZones(cfg)).To(Succeed())

			Expect(ctrl.Zones).To(HaveLen(0))
		})

		It("Should add zones with virtual outputs", func() {
			cfg["foo"] = config.ZoneConfig{Virtual: true}
			cfg["bar"] = config.ZoneConfig{Virtual: true}

			Expect(ctrl.SetupZones(cfg)).To(Succeed())

			Expect(ctrl.Zones).To(HaveLen(2))

			Expect(ctrl.Zones).To(HaveKey("foo"))
			Expect(ctrl.Zones).To(HaveKey("bar"))
			Expect(ctrl.Zones["foo"].Out.Id()).To(Equal("foo"))
			Expect(ctrl.Zones["bar"].Out.Id()).To(Equal("bar"))
		})

		Describe("configuring a thermostat", func() {
			BeforeEach(func() {
				cfg["foo"] = config.ZoneConfig{
					Virtual: true,
					Thermostat: &config.ThermostatConfig{
						Sensor:        "bar",
						DefaultTarget: 18500,
					},
				}
			})
			It("should add a thermostat when configured", func() {
				ctrl.AddSensor("bar", sensor.NewPushSensor("bar"))

				Expect(ctrl.SetupZones(cfg)).To(Succeed())
				Expect(ctrl.Zones).To(HaveLen(1))
				Expect(ctrl.Zones).To(HaveKey("foo"))

				Expect(ctrl.Zones["foo"].Thermostat).NotTo(BeNil())
			})

			It("errors when the given sensor doesn't exist", func() {
				Expect(ctrl.SetupZones(cfg)).NotTo(Succeed())
			})
		})

		It("Should restore the state of the zones", func() {
			writeJSONToFile(DataDir+"/ch.json", map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 45, "action": "Off"},
				},
			})
			cfg["ch"] = config.ZoneConfig{Virtual: true}

			Expect(ctrl.SetupZones(cfg)).To(Succeed())

			Expect(ctrl.Zones).To(HaveLen(1))
			events := ctrl.Zones["ch"].Scheduler.ReadEvents()
			Expect(events).To(HaveLen(2))
		})

		It("Should start the scheduler for the zone", func() {
			cfg["ch"] = config.ZoneConfig{Virtual: true}
			Expect(ctrl.SetupZones(cfg)).To(Succeed())

			Expect(ctrl.Zones).To(HaveLen(1))

			Expect(ctrl.Zones["ch"].Scheduler.Running()).To(BeTrue())
		})

		It("Should add real outputs with correct pin", func() {
			cfg["foo"] = config.ZoneConfig{GPIOPin: 10}
			cfg["bar"] = config.ZoneConfig{GPIOPin: 47}

			Expect(ctrl.SetupZones(cfg)).To(Succeed())

			Expect(ctrl.Zones).To(HaveLen(2))

			Expect(ctrl.Zones).To(HaveKey("foo"))
			Expect(ctrl.Zones).To(HaveKey("bar"))
			Expect(ctrl.Zones["foo"].Out.Id()).To(Equal("foo-gpio10"))
			Expect(ctrl.Zones["bar"].Out.Id()).To(Equal("bar-gpio47"))
		})
	})
})
