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

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/zone"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	log.SetOutput(ioutil.Discard)
	RunSpecs(t, "Main")
}

type testZoneAdder struct {
	Zones []*zone.Zone
}

func (t *testZoneAdder) AddZone(z *zone.Zone) {
	t.Zones = append(t.Zones, z)
}

var _ = Describe("Reading zones from cmdline", func() {
	var (
		srv *testZoneAdder
	)

	BeforeEach(func() {
		zone.DataDir, _ = ioutil.TempDir("", "heating-controller-test")

		srv = &testZoneAdder{make([]*zone.Zone, 0)}
		output_New = func(id string, pin int) (output.Output, error) {
			out := output.Virtual(fmt.Sprintf("%s-gpio%d", id, pin))
			return out, nil
		}
	})
	AfterEach(func() {
		for _, z := range srv.Zones {
			z.Scheduler.Stop()
		}
		os.RemoveAll(zone.DataDir)
	})

	It("Should do nothing with a blank list of zones", func() {
		err := setupZones("", srv)
		Expect(err).To(BeNil())

		Expect(srv.Zones).To(HaveLen(0))
	})

	It("Should add zones with virtual outputs", func() {
		err := setupZones("foo:v,bar:v", srv)
		Expect(err).To(BeNil())

		Expect(srv.Zones).To(HaveLen(2))

		Expect(srv.Zones[0].ID).To(Equal("foo"))
		Expect(srv.Zones[1].ID).To(Equal("bar"))
		Expect(srv.Zones[0].Out.Id()).To(Equal("foo"))
		Expect(srv.Zones[1].Out.Id()).To(Equal("bar"))
	})

	It("Should restore the state of the zones", func() {
		writeJSONToFile(zone.DataDir+"/ch.json", map[string]interface{}{
			"events": []map[string]interface{}{
				{"hour": 6, "min": 30, "action": "On"},
				{"hour": 7, "min": 45, "action": "Off"},
			},
		})
		err := setupZones("ch:v", srv)
		Expect(err).NotTo(HaveOccurred())

		Expect(srv.Zones).To(HaveLen(1))
		events := srv.Zones[0].Scheduler.ReadEvents()
		Expect(events).To(HaveLen(2))
	})

	It("Should start the scheduler for the zone", func() {
		err := setupZones("foo:v", srv)
		Expect(err).To(BeNil())

		Expect(srv.Zones).To(HaveLen(1))

		Expect(srv.Zones[0].Scheduler.Running()).To(BeTrue())
	})

	It("Should add real outputs with correct pin", func() {
		err := setupZones("foo:10,bar:47", srv)
		Expect(err).To(BeNil())

		Expect(srv.Zones).To(HaveLen(2))

		Expect(srv.Zones[0].ID).To(Equal("foo"))
		Expect(srv.Zones[1].ID).To(Equal("bar"))
		Expect(srv.Zones[0].Out.Id()).To(Equal("foo-gpio10"))
		Expect(srv.Zones[1].Out.Id()).To(Equal("bar-gpio47"))
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
