package output

//go:generate counterfeiter -o gpiofakes/fake_pin.go ../vendor/github.com/alext/gpio Pin

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/gpio"
	"github.com/alext/heating-controller/output/gpiofakes"
)

func TestOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Output")
}

var _ = Describe("constructing the gpio instance", func() {

	It("should open the given gpio pin in out mode", func() {
		pinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
			Expect(pin).To(Equal(12))
			Expect(mode).To(Equal(gpio.ModeOutput))
			return new(gpiofakes.FakePin), nil
		}
		_, err := New("foo", 12)
		Expect(err).To(BeNil())
	})

	It("should return any error raised when opening", func() {
		pinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
			return nil, errors.New("computer says no")
		}
		out, err := New("foo", 12)
		Expect(err.Error()).To(Equal("computer says no"))
		Expect(out).To(BeNil())
	})
})

var _ = Describe("Heating control output", func() {
	var (
		fakePin *gpiofakes.FakePin
		output  Output
	)

	BeforeEach(func() {
		fakePin = new(gpiofakes.FakePin)
		pinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
			return fakePin, nil
		}
	})

	JustBeforeEach(func() {
		output, _ = New("foo", 22)
	})

	It("should return the id", func() {
		Expect(output.Id()).To(Equal("foo"))
	})

	Describe("reading the output state", func() {
		It("should return true if the gpio value is 1", func() {
			fakePin.GetReturns(true, nil)

			a, e := output.Active()
			Expect(a).To(BeTrue())
			Expect(e).To(BeNil())
		})

		It("should return false otherwise", func() {
			fakePin.GetReturns(false, nil)

			a, e := output.Active()
			Expect(a).To(BeFalse())
			Expect(e).To(BeNil())
		})

		It("should handle errors", func() {
			err := errors.New("computer says no")
			fakePin.GetReturns(false, err)

			_, e := output.Active()
			Expect(e).To(Equal(err))
		})
	})

	Describe("Activating the output", func() {
		It("should set the gpio pin", func() {
			Expect(output.Activate()).To(BeNil())

			Expect(fakePin.SetCallCount()).To(Equal(1))
		})

		It("should handle errors", func() {
			err := errors.New("computer says no")
			fakePin.SetReturns(err)

			Expect(output.Activate()).To(Equal(err))
		})
	})

	Describe("De-activating the output", func() {
		It("should clear the gpio pin", func() {
			Expect(output.Deactivate()).To(BeNil())

			Expect(fakePin.ClearCallCount()).To(Equal(1))
		})

		It("should handle errors", func() {
			err := errors.New("computer says no")
			fakePin.ClearReturns(err)

			Expect(output.Deactivate()).To(Equal(err))
		})
	})

	Describe("closing the output", func() {
		It("should close the gpio pin", func() {
			Expect(output.Close()).To(BeNil())

			Expect(fakePin.CloseCallCount()).To(Equal(1))
		})

		It("should return any error received closing the gpio pin", func() {
			err := errors.New("Boom!")
			fakePin.CloseReturns(err)

			Expect(output.Close()).To(Equal(err))
		})
	})
})
