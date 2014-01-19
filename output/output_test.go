package output

import (
	"code.google.com/p/gomock/gomock"
	"errors"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/thirdparty/gomocktestreporter"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/mock_gpio"
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
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(gomocktestreporter.New())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should open the given gpio pin in out mode", func() {
		pinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
			Expect(pin).To(Equal(12))
			Expect(mode).To(Equal(gpio.ModeOutput))
			return mock_gpio.NewMockPin(mockCtrl), nil
		}
		_, err := NewOutput("foo", 12)
		Expect(err).To(BeNil())
	})

	It("should return any error raised when opening", func() {
		pinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
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
		pinOpener = func(pin int, mode gpio.Mode) (gpio.Pin, error) {
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
		var errStub *gomock.Call
		BeforeEach(func() {
			errStub = mockPin.EXPECT().Err().Return(nil)
		})

		It("should return true if the gpio value is 1", func() {
			mockPin.EXPECT().Get().Return(true)

			a, e := output.Active()
			Expect(a).To(BeTrue())
			Expect(e).To(BeNil())
		})

		It("should return false otherwise", func() {
			mockPin.EXPECT().Get().Return(false)

			a, e := output.Active()
			Expect(a).To(BeFalse())
			Expect(e).To(BeNil())
		})

		It("should handle errors", func() {
			err := errors.New("computer says no")
			errStub.Return(err)
			mockPin.EXPECT().Get().Return(false)

			_, e := output.Active()
			Expect(e).To(Equal(err))
		})
	})

	Describe("Activating the output", func() {
		var errStub *gomock.Call
		BeforeEach(func() {
			errStub = mockPin.EXPECT().Err().Return(nil)
			mockPin.EXPECT().Set()
		})

		It("should set the gpio pin", func() {
			Expect(output.Activate()).To(BeNil())
		})

		It("should handle errors", func() {
			err := errors.New("computer says no")
			errStub.Return(err)

			Expect(output.Activate()).To(Equal(err))
		})
	})

	Describe("De-activating the output", func() {
		var errStub *gomock.Call
		BeforeEach(func() {
			errStub = mockPin.EXPECT().Err().Return(nil)
			mockPin.EXPECT().Clear()
		})

		It("should clear the gpio pin", func() {
			Expect(output.Deactivate()).To(BeNil())
		})

		It("should handle errors", func() {
			err := errors.New("computer says no")
			errStub.Return(err)

			Expect(output.Deactivate()).To(Equal(err))
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
