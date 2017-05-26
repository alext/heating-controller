package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/config"
	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/output"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	log.SetOutput(ioutil.Discard)
	RunSpecs(t, "Main")
}

type testZoneAdder struct {
	Zones map[string]*controller.Zone
}

func (t *testZoneAdder) AddZone(z *controller.Zone) {
	t.Zones[z.ID] = z
}

var _ = Describe("Setting up zones from config file", func() {
	var (
		srv *testZoneAdder
		cfg map[string]config.ZoneConfig
	)

	BeforeEach(func() {
		controller.DataDir, _ = ioutil.TempDir("", "heating-controller-test")

		cfg = make(map[string]config.ZoneConfig)
		srv = &testZoneAdder{make(map[string]*controller.Zone)}
		output_New = func(id string, pin int) (output.Output, error) {
			out := output.Virtual(fmt.Sprintf("%s-gpio%d", id, pin))
			return out, nil
		}
	})
	AfterEach(func() {
		for _, z := range srv.Zones {
			z.Scheduler.Stop()
		}
		os.RemoveAll(controller.DataDir)
	})

	It("Should do nothing with a blank list of zones", func() {
		Expect(setupZones(cfg, srv)).To(Succeed())

		Expect(srv.Zones).To(HaveLen(0))
	})

	It("Should add zones with virtual outputs", func() {
		cfg["foo"] = config.ZoneConfig{Virtual: true}
		cfg["bar"] = config.ZoneConfig{Virtual: true}

		Expect(setupZones(cfg, srv)).To(Succeed())

		Expect(srv.Zones).To(HaveLen(2))

		Expect(srv.Zones).To(HaveKey("foo"))
		Expect(srv.Zones).To(HaveKey("bar"))
		Expect(srv.Zones["foo"].Out.Id()).To(Equal("foo"))
		Expect(srv.Zones["bar"].Out.Id()).To(Equal("bar"))
	})

	It("should add a thermostat when configured", func() {
		cfg["foo"] = config.ZoneConfig{
			Virtual: true,
			Thermostat: &config.ThermostatConfig{
				SensorURL:     "http://foo.example.com/foo",
				DefaultTarget: 18500,
			},
		}

		Expect(setupZones(cfg, srv)).To(Succeed())
		Expect(srv.Zones).To(HaveLen(1))
		Expect(srv.Zones).To(HaveKey("foo"))

		Expect(srv.Zones["foo"].Thermostat).NotTo(BeNil())
	})

	It("Should restore the state of the zones", func() {
		writeJSONToFile(controller.DataDir+"/ch.json", map[string]interface{}{
			"events": []map[string]interface{}{
				{"hour": 6, "min": 30, "action": "On"},
				{"hour": 7, "min": 45, "action": "Off"},
			},
		})
		cfg["ch"] = config.ZoneConfig{Virtual: true}

		Expect(setupZones(cfg, srv)).To(Succeed())

		Expect(srv.Zones).To(HaveLen(1))
		events := srv.Zones["ch"].Scheduler.ReadEvents()
		Expect(events).To(HaveLen(2))
	})

	It("Should start the scheduler for the zone", func() {
		cfg["ch"] = config.ZoneConfig{Virtual: true}
		Expect(setupZones(cfg, srv)).To(Succeed())

		Expect(srv.Zones).To(HaveLen(1))

		Expect(srv.Zones["ch"].Scheduler.Running()).To(BeTrue())
	})

	It("Should add real outputs with correct pin", func() {
		cfg["foo"] = config.ZoneConfig{GPIOPin: 10}
		cfg["bar"] = config.ZoneConfig{GPIOPin: 47}

		Expect(setupZones(cfg, srv)).To(Succeed())

		Expect(srv.Zones).To(HaveLen(2))

		Expect(srv.Zones).To(HaveKey("foo"))
		Expect(srv.Zones).To(HaveKey("bar"))
		Expect(srv.Zones["foo"].Out.Id()).To(Equal("foo-gpio10"))
		Expect(srv.Zones["bar"].Out.Id()).To(Equal("bar-gpio47"))
	})
})

func writeJSONToFile(filename string, data interface{}) {
	file, err := os.Create(filename)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	defer file.Close()

	b, err := json.MarshalIndent(data, "", "  ")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	_, err = file.Write(b)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}
