package sensor

import (
	"time"

	"github.com/alext/heating-controller/units"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("a push sensor", func() {

	It("returns the deviceID", func() {
		s := NewPushSensor("something")
		Expect(s.DeviceId()).To(Equal("something"))
	})

	It("has a reasonable initial value", func() {
		s := NewPushSensor("something")
		temp, updated := s.Read()
		Expect(temp).To(BeEquivalentTo(initialValue))
		var zeroTime time.Time
		Expect(updated).To(Equal(zeroTime))
	})

	It("saves and returns the set temerature", func() {
		s := &pushSensor{}
		now := time.Now()
		s.Set(1234, now)
		temp, updated := s.Read()
		Expect(temp).To(BeEquivalentTo(1234))
		Expect(updated).To(Equal(now))
	})

	Describe("subscribing to updates", func() {
		It("allows subscribing to updates", func() {
			s := NewPushSensor("something")
			ch := s.Subscribe()

			s.Set(1234, time.Now())
			Eventually(ch).Should(Receive(Equal(units.Temperature(1234))))
		})

		It("allows multiple subscribers", func() {
			s := NewPushSensor("something")
			ch1 := s.Subscribe()
			ch2 := s.Subscribe()

			s.Set(1234, time.Now())
			Eventually(ch1).Should(Receive(Equal(units.Temperature(1234))))
			Eventually(ch2).Should(Receive(Equal(units.Temperature(1234))))

			ch3 := s.Subscribe()
			s.Set(12345, time.Now())
			Eventually(ch1).Should(Receive(Equal(units.Temperature(12345))))
			Eventually(ch2).Should(Receive(Equal(units.Temperature(12345))))
			Eventually(ch3).Should(Receive(Equal(units.Temperature(12345))))
		})
	})
})
