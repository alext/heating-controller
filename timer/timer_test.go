package timer

import (
	"testing"
	"time"

	"code.google.com/p/gomock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output/mock_output"
)

func TestOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Timer")
}

var _ = Describe("a basic timer", func() {
	var (
		mockCtrl    *gomock.Controller
		output      *mock_output.MockOutput
		mockNow     time.Time
		timer       Timer
		afterParam  time.Duration
		afterCh     chan time.Time
		afterNotify chan bool
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		output = mock_output.NewMockOutput(mockCtrl)
		timer = New(output)

		mockNow = time.Now()
		time_Now = func() time.Time {
			return mockNow
		}

		afterNotify = make(chan bool, 1)
		time_After = func(d time.Duration) <-chan time.Time {
			afterParam = d
			afterCh = make(chan time.Time, 1)

			// Notify the channel, but don't block if nothing is listening.
			select {
			case afterNotify <- true:
			default:
			}

			return afterCh
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("firing events on the regular schedule", func() {

		It("should activate the output at 6:30", func() {
			mockNow = todayAt(6, 20, 0)

			timer.Start()
			<-afterNotify

			expectedDuration, _ := time.ParseDuration("10m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Activate().Return(nil)
			mockNow = todayAt(6, 30, 0)
			afterCh <- mockNow
			<-afterNotify
			timer.Stop()
		})

		It("should deactivate the output at 7:30", func() {
			mockNow = todayAt(7, 20, 0)

			timer.Start()
			<-afterNotify

			expectedDuration, _ := time.ParseDuration("10m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Deactivate().Return(nil)
			mockNow = todayAt(7, 30, 0)
			afterCh <- mockNow
			<-afterNotify
			timer.Stop()
		})

		It("should activate the output at 17:00", func() {
			mockNow = todayAt(16, 45, 0)

			timer.Start()
			<-afterNotify

			expectedDuration, _ := time.ParseDuration("15m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Activate().Return(nil)
			mockNow = todayAt(17, 0, 0)
			afterCh <- mockNow
			<-afterNotify
			timer.Stop()
		})

		It("should deactivate the output at 21:00", func() {
			mockNow = todayAt(20, 0, 0)

			timer.Start()
			<-afterNotify

			expectedDuration, _ := time.ParseDuration("60m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Deactivate().Return(nil)
			mockNow = todayAt(21, 0, 0)
			afterCh <- mockNow
			<-afterNotify
			timer.Stop()
		})
	})
})

func todayAt(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local)
}
