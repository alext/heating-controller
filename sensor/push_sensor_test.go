package sensor

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("a push sensor", func() {

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
})
