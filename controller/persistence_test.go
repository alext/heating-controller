package controller

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/thermostat/thermostatfakes"
)

var _ = Describe("persisting a zone's state", func() {
	var (
		tempDataDir string
		z           *Zone
	)

	BeforeEach(func() {
		tempDataDir, _ = ioutil.TempDir("", "persistence_test")
		DataDir = tempDataDir

		z = NewZone("ch", output.Virtual("something"))
	})
	AfterEach(func() {
		os.RemoveAll(tempDataDir)
	})

	Describe("saving zone state", func() {

		It("should save an empty scheduler event list as JSON", func() {
			Expect(z.Save()).To(Succeed())
			data := readFile(filepath.Join(tempDataDir, "ch.json"))
			Expect(data).To(MatchJSON(`{"events":[]}`))
		})

		It("should save the scheduler events to the file", func() {
			z.AddEvent(Event{Hour: 6, Min: 30, Action: TurnOn})
			z.AddEvent(Event{Hour: 7, Min: 45, Action: TurnOff})

			Expect(z.Save()).To(Succeed())

			data := readFile(filepath.Join(tempDataDir, "ch.json"))
			expected, _ := json.Marshal(map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 45, "action": "Off"},
				},
			})
			Expect(data).To(MatchJSON(expected))
		})

		It("should save the thermostat target", func() {
			t := new(thermostatfakes.FakeThermostat)
			t.TargetReturns(18500)
			z.Thermostat = t
			Expect(z.Save()).To(Succeed())
			data := readFile(filepath.Join(tempDataDir, "ch.json"))
			Expect(data).To(MatchJSON(`{"events":[],"thermostat_target":18500}`))
		})
	})

	Describe("restoring zone state", func() {
		It("should load an empty scheduler event list", func() {
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{"events": []interface{}{}})
			Expect(z.Restore()).To(Succeed())

			Expect(z.ReadEvents()).To(HaveLen(0))
		})

		It("should load the scheduler events from the file", func() {
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 45, "action": "Off"},
				},
			})

			Expect(z.Restore()).To(Succeed())

			events := z.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events[0]).To(Equal(Event{Hour: 6, Min: 30, Action: TurnOn}))
			Expect(events[1]).To(Equal(Event{Hour: 7, Min: 45, Action: TurnOff}))
		})

		It("should treat a non-existent data file the same as a file with an empty scheduler event list", func() {
			Expect(z.Restore()).To(Succeed())

			Expect(z.ReadEvents()).To(HaveLen(0))
		})

		It("should skip over any invalid scheduler events in the file", func() {
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 75, "action": "Off"}, // Invalid minute 75
					{"hour": 16, "min": 30, "action": "On"},
					{"hour": 18, "min": 40, "action": "Off"},
				},
			})

			Expect(z.Restore()).To(Succeed())

			events := z.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events[0]).To(Equal(Event{Hour: 6, Min: 30, Action: TurnOn}))
			Expect(events[1]).To(Equal(Event{Hour: 16, Min: 30, Action: TurnOn}))
			Expect(events[2]).To(Equal(Event{Hour: 18, Min: 40, Action: TurnOff}))
		})

		It("should restore the thermostat target", func() {
			t := new(thermostatfakes.FakeThermostat)
			z.Thermostat = t
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{
				"thermostat_target": 19000,
			})

			Expect(z.Restore()).To(Succeed())

			Expect(t.SetCallCount()).To(Equal(1))
			Expect(t.SetArgsForCall(0)).To(BeNumerically("==", 19000))
		})

		It("should leave the thermostat target unchanged if not present in the file", func() {
			t := new(thermostatfakes.FakeThermostat)
			z.Thermostat = t
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{})

			Expect(z.Restore()).To(Succeed())

			Expect(t.SetCallCount()).To(Equal(0))
		})

		It("should ignore a thermostat target in the file for a zone with no thermostat configured", func() {
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{
				"thermostat_target": 19000,
			})

			Expect(z.Restore()).To(Succeed())
		})
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

func readFile(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return data
}
