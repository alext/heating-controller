package sensor

import (
	"os"
	"time"

	"github.com/alext/heating-controller/units"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/afero"
)

const testSensorID = "28-0123456789ab"

const sampleData1 = `37 01 4b 46 7f ff 09 10 26 : crc=26 YES
37 01 4b 46 7f ff 09 10 26 t=19437
`
const sampleData2 = `21 01 4b 46 7f ff 0f 10 4b : crc=4b YES
21 01 4b 46 7f ff 0f 10 4b t=18062
`

type dummyTicker struct {
	duration time.Duration
	C        chan time.Time
	notify   chan struct{}
	stopped  bool
}

func (t dummyTicker) Channel() <-chan time.Time {
	select {
	case t.notify <- struct{}{}:
	default:
	}
	return t.C
}
func (t *dummyTicker) Stop() {
	t.stopped = true
}

var _ = Describe("a w1 sensor", func() {
	var (
		tkr       *dummyTicker
		tkrNotify chan struct{}
	)

	BeforeEach(func() {
		fs = &afero.MemMapFs{}
		tkrNotify = make(chan struct{}, 1)

		newTicker = func(d time.Duration) ticker {
			tkr = &dummyTicker{
				duration: d,
				C:        make(chan time.Time, 1),
				notify:   tkrNotify,
			}
			return tkr
		}
	})

	Describe("constructing a sensor", func() {
		var (
			sensor Sensor
		)

		BeforeEach(func() {
			populateValueFile(testSensorID, sampleData1)
			sensor = NewW1Sensor("foo", testSensorID)
		})
		AfterEach(func() {
			if s, ok := sensor.(*w1Sensor); ok {
				s.Close()
			}
		})

		It("returns the ID", func() {
			Expect(sensor.ID()).To(Equal(testSensorID))
		})

		It("should read the initial temperature", func() {
			temperature, _ := sensor.Read()
			Expect(temperature).To(BeEquivalentTo(19437))
		})

		It("should start a ticker to poll the temperature every minute", func(done Done) {
			<-tkrNotify
			Expect(tkr.duration).To(Equal(time.Minute))

			close(done)
		})

		It("should update the temperature on each tick", func(done Done) {
			<-tkrNotify
			temperature, _ := sensor.Read()
			Expect(temperature).To(BeEquivalentTo(19437))

			populateValueFile(testSensorID, sampleData2)
			tkr.C <- time.Now()
			<-tkrNotify
			temperature, _ = sensor.Read()
			Expect(temperature).To(BeEquivalentTo(18062))

			populateValueFile(testSensorID, sampleData1)
			tkr.C <- time.Now()
			<-tkrNotify
			temperature, _ = sensor.Read()
			Expect(temperature).To(BeEquivalentTo(19437))

			close(done)
		})

		It("should handle negative temperatures", func(done Done) {
			<-tkrNotify

			negativeData := `f6 ff 55 00 7f ff 0c 10 47 : crc=47 YES
f6 ff 55 00 7f ff 0c 10 47 t=-625`
			populateValueFile(testSensorID, negativeData)
			tkr.C <- time.Now()
			<-tkrNotify
			temperature, _ := sensor.Read()
			Expect(temperature).To(BeEquivalentTo(-625))

			close(done)
		})

		It("should track when the temperature was last updated", func(done Done) {
			<-tkrNotify
			populateValueFile(testSensorID, sampleData2)

			tickTime := time.Now().Add(-15 * time.Minute)
			tkr.C <- tickTime
			<-tkrNotify
			_, updatedAt := sensor.Read()
			Expect(updatedAt).To(Equal(tickTime))

			populateValueFile(testSensorID, sampleData1)
			tickTime = time.Now().Add(-3 * time.Minute)
			tkr.C <- tickTime
			<-tkrNotify
			_, updatedAt = sensor.Read()
			Expect(updatedAt).To(Equal(tickTime))

			close(done)
		})

		It("allows subscribing to updates", func() {
			<-tkrNotify

			ch := sensor.Subscribe()
			tkr.C <- time.Now()
			<-tkrNotify
			Eventually(ch).Should(Receive(Equal(units.Temperature(19437))))

			populateValueFile(testSensorID, sampleData2)
			tkr.C <- time.Now()
			<-tkrNotify
			Eventually(ch).Should(Receive(Equal(units.Temperature(18062))))
		})
	})

	Describe("closing a sensor", func() {
		var (
			sensor *w1Sensor
		)

		BeforeEach(func() {
			populateValueFile(testSensorID, sampleData1)
			sensor = NewW1Sensor("foo", testSensorID).(*w1Sensor)
		})

		It("should stop the ticker", func(done Done) {
			<-tkrNotify

			sensor.Close()
			Expect(tkr.stopped).To(BeTrue())

			close(done)
		})
	})
})

func populateValueFile(id, contents string) {
	valueFilePath := w1DevicesPath + id + "/w1_slave"
	file, err := fs.OpenFile(valueFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	_, err = file.Write([]byte(contents))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}
