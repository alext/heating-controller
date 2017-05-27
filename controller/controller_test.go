package controller

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/config"
	"github.com/alext/heating-controller/output"
)

var _ = Describe("Controller", func() {

	Describe("Setting up the controller", func() {
		var (
			ctrl *Controller
			cfg  *config.Config
		)
		BeforeEach(func() {
			var err error
			DataDir, err = ioutil.TempDir("", "heating-controller-test")
			Expect(err).NotTo(HaveOccurred())

			cfg = config.New()
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

		It("should do nothing with no sensors or zones", func() {
			Expect(ctrl.Setup(cfg)).To(Succeed())

			Expect(ctrl.Sensors).To(HaveLen(0))
			Expect(ctrl.Zones).To(HaveLen(0))
		})

		Describe("setting up sensors", func() {
			It("should add sensors based on the config", func() {
				cfg.Sensors["foo"] = config.SensorConfig{
					Type: "push",
					ID:   "1234",
				}
				cfg.Sensors["bar"] = config.SensorConfig{
					Type: "push",
					ID:   "2345",
				}

				Expect(ctrl.Setup(cfg)).To(Succeed())

				Expect(ctrl.Sensors).To(HaveLen(2))
			})

			It("should return an error if setting up a sensor fails", func() {
				cfg.Sensors["foo"] = config.SensorConfig{
					Type: "non-existent",
					ID:   "1234",
				}
				Expect(ctrl.Setup(cfg)).NotTo(Succeed())
			})
		})

		Describe("Setting up zones", func() {
			It("Should add zones with virtual outputs", func() {
				cfg.Zones["foo"] = config.ZoneConfig{Virtual: true}
				cfg.Zones["bar"] = config.ZoneConfig{Virtual: true}

				Expect(ctrl.Setup(cfg)).To(Succeed())

				Expect(ctrl.Zones).To(HaveLen(2))

				Expect(ctrl.Zones).To(HaveKey("foo"))
				Expect(ctrl.Zones).To(HaveKey("bar"))
				Expect(ctrl.Zones["foo"].Out.Id()).To(Equal("foo"))
				Expect(ctrl.Zones["bar"].Out.Id()).To(Equal("bar"))
			})

			Describe("configuring a thermostat", func() {
				BeforeEach(func() {
					cfg.Zones["foo"] = config.ZoneConfig{
						Virtual: true,
						Thermostat: &config.ThermostatConfig{
							Sensor:        "bar",
							DefaultTarget: 18500,
						},
					}
				})
				It("should add a thermostat when configured", func() {
					cfg.Sensors["bar"] = config.SensorConfig{
						Type: "push",
						ID:   "bar",
					}

					Expect(ctrl.Setup(cfg)).To(Succeed())

					Expect(ctrl.Zones).To(HaveLen(1))
					Expect(ctrl.Zones).To(HaveKey("foo"))

					Expect(ctrl.Zones["foo"].Thermostat).NotTo(BeNil())
				})

				It("errors when the given sensor doesn't exist", func() {
					Expect(ctrl.Setup(cfg)).NotTo(Succeed())
				})
			})

			It("Should restore the state of the zones", func() {
				writeJSONToFile(DataDir+"/ch.json", map[string]interface{}{
					"events": []map[string]interface{}{
						{"hour": 6, "min": 30, "action": "On"},
						{"hour": 7, "min": 45, "action": "Off"},
					},
				})
				cfg.Zones["ch"] = config.ZoneConfig{Virtual: true}

				Expect(ctrl.Setup(cfg)).To(Succeed())

				Expect(ctrl.Zones).To(HaveLen(1))
				events := ctrl.Zones["ch"].Scheduler.ReadEvents()
				Expect(events).To(HaveLen(2))
			})

			It("Should start the scheduler for the zone", func() {
				cfg.Zones["ch"] = config.ZoneConfig{Virtual: true}
				Expect(ctrl.Setup(cfg)).To(Succeed())

				Expect(ctrl.Zones).To(HaveLen(1))

				Expect(ctrl.Zones["ch"].Scheduler.Running()).To(BeTrue())
			})

			It("Should add real outputs with correct pin", func() {
				cfg.Zones["foo"] = config.ZoneConfig{GPIOPin: 10}
				cfg.Zones["bar"] = config.ZoneConfig{GPIOPin: 47}

				Expect(ctrl.Setup(cfg)).To(Succeed())

				Expect(ctrl.Zones).To(HaveLen(2))

				Expect(ctrl.Zones).To(HaveKey("foo"))
				Expect(ctrl.Zones).To(HaveKey("bar"))
				Expect(ctrl.Zones["foo"].Out.Id()).To(Equal("foo-gpio10"))
				Expect(ctrl.Zones["bar"].Out.Id()).To(Equal("bar-gpio47"))
			})
		})
	})
})
