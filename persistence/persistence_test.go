package persistence_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/persistence"
	"github.com/alext/heating-controller/scheduler"
)

func TestOutput(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Scheduler")
}

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

var _ = Describe("persisting scheduler events", func() {
	var (
		tempDataDir string
		eventHolder *EventHolder
	)

	BeforeEach(func() {
		tempDataDir, _ = ioutil.TempDir("", "persistence_test")
		persistence.DataDir = tempDataDir

		eventHolder = &EventHolder{events: make([]scheduler.Event, 0, 0)}
	})
	AfterEach(func() {
		os.RemoveAll(tempDataDir)
	})

	Describe("saving event list", func() {

		It("should save an empty event list as JSON", func() {
			Expect(persistence.SaveEvents("ch", eventHolder)).To(Succeed())
			data := readFile(tempDataDir + "/ch.json")
			Expect(data).To(MatchJSON(`{"events":[]}`))
		})

		It("should save the events to the file", func() {
			eventHolder.AddEvent(scheduler.Event{Hour: 6, Min: 30, Action: scheduler.TurnOn})
			eventHolder.AddEvent(scheduler.Event{Hour: 7, Min: 45, Action: scheduler.TurnOff})

			Expect(persistence.SaveEvents("ch", eventHolder)).To(Succeed())

			data := readFile(tempDataDir + "/ch.json")
			expected, _ := json.Marshal(map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 45, "action": "Off"},
				},
			})
			Expect(data).To(MatchJSON(expected))
		})
	})

	Describe("reading events", func() {
		It("should load an empty event list", func() {
			writeJSONToFile(tempDataDir+"/ch.json", map[string]interface{}{"events": []interface{}{}})
			Expect(persistence.LoadEvents("ch", eventHolder)).To(Succeed())

			Expect(eventHolder.ReadEvents()).To(HaveLen(0))
		})

		It("should load the events from the file", func() {
			writeJSONToFile(tempDataDir+"/ch.json", map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 6, "min": 30, "action": "On"},
					{"hour": 7, "min": 45, "action": "Off"},
				},
			})

			Expect(persistence.LoadEvents("ch", eventHolder)).To(Succeed())

			Expect(eventHolder.events).To(HaveLen(2))
			Expect(eventHolder.events[0]).To(Equal(scheduler.Event{Hour: 6, Min: 30, Action: scheduler.TurnOn}))
			Expect(eventHolder.events[1]).To(Equal(scheduler.Event{Hour: 7, Min: 45, Action: scheduler.TurnOff}))
		})

		It("should treat a non-existent data file the same as a file with an empty event list", func() {
			Expect(persistence.LoadEvents("ch", eventHolder)).To(Succeed())

			Expect(eventHolder.ReadEvents()).To(HaveLen(0))
		})
	})
})

var _ = Describe("scheduler implementing persistence interfaces", func() {
	var theScheduler scheduler.Scheduler

	BeforeEach(func() {
		theScheduler = scheduler.New("foo", func(scheduler.Action) {})
	})

	It("should implement the EventAdder interface", func() {
		// Will fail to build if the interfaces don't match
		var _ persistence.EventAdder = theScheduler
	})

	It("should implement the EventReader interface", func() {
		// Will fail to build if the interfaces don't match
		var _ persistence.EventReader = theScheduler
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
