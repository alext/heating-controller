package output_test

import (
	"code.google.com/p/gomock/gomock"
	"errors"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/thirdparty/gomocktestreporter"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/mock_gpio"
	. "github.com/alext/heating-controller/output"
	"github.com/davecheney/gpio"
	"testing"
)

func TestOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Output")
}

var _ = Describe("constructing the gpio instance", func() {
	var (
		mockCtrl *gomock.Controller
		//output Output
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(gomocktestreporter.New())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should open the given gpio pin in out mode", func() {
		PinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
			Expect(pin).To(Equal(12))
			Expect(mode).To(Equal(gpio.ModeOutput))
			return mock_gpio.NewMockPin(mockCtrl), nil
		}
		_, err := NewOutput("foo", 12)
		Expect(err).To(BeNil())
	})

	It("should return any error raised when opening", func() {
		PinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
			return nil, errors.New("computer says no")
		}
		out, err := NewOutput("foo", 12)
		Expect(err.Error()).To(Equal("computer says no"))
		Expect(out).To(BeNil())
	})
})

var _ = Describe("Heating control output", func() {
	var (
		mockCtrl *gomock.Controller
		mockPin  *mock_gpio.MockPin
		output   Output
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(gomocktestreporter.New())
		mockPin = mock_gpio.NewMockPin(mockCtrl)
		PinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
			return mockPin, nil
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	JustBeforeEach(func() {
		output, _ = NewOutput("foo", 22)
	})

	It("should return the id", func() {
		Expect(output.Id()).To(Equal("foo"))
	})

	Describe("reading the output state", func() {
		It("should return true if the gpio value is 1", func() {
			mockPin.EXPECT().Get().Return(true)

			Expect(output.Active()).To(BeTrue())
		})

		It("should return false otherwise", func() {
			mockPin.EXPECT().Get().Return(false)

			Expect(output.Active()).To(BeFalse())
		})
	})

	Describe("Activating the output", func() {
		It("should set the gpio pin", func() {
			mockPin.EXPECT().Set()
			output.Activate()
		})
	})

	Describe("De-activating the output", func() {
		It("should clear the gpio pin", func() {
			mockPin.EXPECT().Clear()
			output.Deactivate()
		})
	})

	Describe("closing the output", func() {
		It("should close the gpio pin", func() {
			mockPin.EXPECT().Close().Return(nil)

			Expect(output.Close()).To(BeNil())
		})

		It("should return any error received closing the gpio pin", func() {
			err := errors.New("Boom!")
			mockPin.EXPECT().Close().Return(err)

			Expect(output.Close()).To(Equal(err))
		})
	})
})
