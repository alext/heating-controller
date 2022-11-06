package sensor_test

import (
	"errors"

	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/sensor/sensorfakes"
	"github.com/alext/heating-controller/units"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type sensorCloser interface {
	Close()
}

var _ = Describe("MQTT Sensor", func() {
	var (
		fakeSubscriber *sensorfakes.FakeTopicSubscriber

		topicCh chan string
		sens    sensor.Sensor
		err     error
	)

	BeforeEach(func() {
		fakeSubscriber = &sensorfakes.FakeTopicSubscriber{}
		topicCh = make(chan string)
		fakeSubscriber.SubscribeReturns(topicCh, nil)
	})
	AfterEach(func() {
		if cl, ok := sens.(sensorCloser); ok {
			cl.Close()
		}
	})

	Describe("constructing a sensor", func() {

		It("sets an initial value and subscribes to the topic", func() {
			sens, err = sensor.NewMQTTSensor("one", "some/topic", fakeSubscriber)
			Expect(err).NotTo(HaveOccurred())

			t, _ := sens.Read()
			Expect(t).To(BeNumerically("==", 21_000))

			Expect(fakeSubscriber.SubscribeCallCount()).To(Equal(1))
			Expect(fakeSubscriber.SubscribeArgsForCall(0)).To(Equal("some/topic"))
		})

		It("returns an error if subscribing errors", func() {
			fakeSubscriber.SubscribeReturns(nil, errors.New("Computer says no"))

			_, err := sensor.NewMQTTSensor("one", "some/topic", fakeSubscriber)
			Expect(err).To(MatchError("Computer says no"))
		})
	})

	Describe("handling temperature updates", func() {

		BeforeEach(func() {
			sens, err = sensor.NewMQTTSensor("one", "some/topic", fakeSubscriber)
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates the temperature value", func() {
			topicCh <- "21.345"

			Eventually(func() units.Temperature {
				t, _ := sens.Read()
				return t
			}).Should(BeNumerically("==", 21_345))
		})

		It("ignores messages that can't be parsed", func() {
			topicCh <- "wibble"

			// Should not update value
			Consistently(func() units.Temperature {
				t, _ := sens.Read()
				return t
			}).Should(BeNumerically("==", 21_000))

			// Should continue to parse future values
			topicCh <- "19.175"
			Eventually(func() units.Temperature {
				t, _ := sens.Read()
				return t
			}).Should(BeNumerically("==", 19_175))
		})
	})
})
