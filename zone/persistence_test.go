package zone

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/scheduler"
)

type EventHolder struct {
	events []scheduler.Event
}

func (eh *EventHolder) ReadEvents() []scheduler.Event {
	return eh.events
}

func (eh *EventHolder) AddEvent(e scheduler.Event) error {
	eh.events = append(eh.events, e)
	return nil
}

var _ = Describe("persisting a zone's state", func() {
	var (
		tempDataDir string
		z           *Zone
	)

	BeforeEach(func() {
		tempDataDir, _ = ioutil.TempDir("", "persistence_test")
		DataDir = tempDataDir

		z = New("ch", output.Virtual("something"))
	})
	AfterEach(func() {
		os.RemoveAll(tempDataDir)
	})

	Describe("saving zone state", func() {

		It("should save an blank zone as JSON", func() {
			Expect(z.Save()).To(Succeed())
			data := readFile(filepath.Join(tempDataDir, "ch.json"))
			expected, _ := json.Marshal(map[string]interface{}{
				"override_mode": "normal",
				"events":        []interface{}{},
			})
			Expect(data).To(MatchJSON(expected))
		})

		It("should save the events to the file", func() {
			z.Scheduler.AddEvent(scheduler.Event{Hour: 6, Min: 30, Action: scheduler.TurnOn})
			z.Scheduler.AddEvent(scheduler.Event{Hour: 7, Min: 45, Action: scheduler.TurnOff})

			Expect(z.Save()).To(Succeed())

			data := readFile(filepath.Join(tempDataDir, "ch.json"))
			expected, _ := json.Marshal(map[string]interface{}{
				"override_mode": "normal",
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 45, "action": "Off"},
				},
			})
			Expect(data).To(MatchJSON(expected))
		})

		It("should include the override mode in the file", func() {
			z.SetOverride(ModeOverrideOn)
			Expect(z.Save()).To(Succeed())
			data := readFile(filepath.Join(tempDataDir, "ch.json"))
			expected, _ := json.Marshal(map[string]interface{}{
				"override_mode": "override_on",
				"events":        []interface{}{},
			})
			Expect(data).To(MatchJSON(expected))
		})
	})

	Describe("restoring zone state", func() {
		It("should load an empty event list", func() {
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{"events": []interface{}{}})
			Expect(z.Restore()).To(Succeed())

			Expect(z.Scheduler.ReadEvents()).To(HaveLen(0))
			Expect(z.overrideMode).To(Equal(ModeNormal))
		})

		It("should load the events from the file", func() {
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 45, "action": "Off"},
				},
			})

			Expect(z.Restore()).To(Succeed())

			events := z.Scheduler.ReadEvents()
			Expect(events).To(HaveLen(2))
			Expect(events[0]).To(Equal(scheduler.Event{Hour: 6, Min: 30, Action: scheduler.TurnOn}))
			Expect(events[1]).To(Equal(scheduler.Event{Hour: 7, Min: 45, Action: scheduler.TurnOff}))
		})

		It("should load default values with a non-existent state file", func() {
			Expect(z.Restore()).To(Succeed())

			Expect(z.overrideMode).To(Equal(ModeNormal))
			Expect(z.Scheduler.ReadEvents()).To(HaveLen(0))
		})

		It("should restore the override mode", func() {
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{
				"override_mode": "override_off",
				"events":        []interface{}{},
			})
			Expect(z.Restore()).To(Succeed())

			Expect(z.overrideMode).To(Equal(ModeOverrideOff))
		})

		It("should skip over any invalid events in the file", func() {
			writeJSONToFile(filepath.Join(tempDataDir, "ch.json"), map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 75, "action": "Off"}, // Invalid minute 75
					{"hour": 16, "min": 30, "action": "On"},
					{"hour": 18, "min": 40, "action": "Off"},
				},
			})

			Expect(z.Restore()).To(Succeed())

			events := z.Scheduler.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events[0]).To(Equal(scheduler.Event{Hour: 6, Min: 30, Action: scheduler.TurnOn}))
			Expect(events[1]).To(Equal(scheduler.Event{Hour: 16, Min: 30, Action: scheduler.TurnOn}))
			Expect(events[2]).To(Equal(scheduler.Event{Hour: 18, Min: 40, Action: scheduler.TurnOff}))
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
