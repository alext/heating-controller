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
		nowCount    int
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
		nowCount = 0
		time_Now = func() time.Time {
			nowCount++
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
		timer.Stop()
		mockCtrl.Finish()
	})

	Describe("starting and stopping the timer", func() {
		It("should not be running when newly created", func() {
			Expect(timer.Running()).To(BeFalse())
		})

		It("should start the timer", func() {
			timer.Start()
			Expect(timer.Running()).To(BeTrue())
		})

		It("should do nothing when attempting to start a running timer", func() {
			timer.Start()
			timer.Start()
			<-afterNotify

			Expect(nowCount).To(Equal(1))
		})

		It("should stop the timer", func() {
			timer.Start()
			timer.Stop()
			Expect(timer.Running()).To(BeFalse())
		})

		It("should do nothing when attempting to stop a non-running timer", func(done Done) {
			timer.Stop()
			close(done)
		}, 0.5)
	})

	Describe("firing events as scheduled", func() {

		BeforeEach(func() {
			timer.AddEntry(6, 30, TurnOn)
			timer.AddEntry(7, 45, TurnOff)
			timer.AddEntry(17, 33, TurnOn)
			timer.AddEntry(21, 12, TurnOff)
		})

		PIt("should fire the given events in order", func() {
			mockNow = todayAt(6, 20, 0)

			timer.Start()
			<-afterNotify

			expectedDuration, _ := time.ParseDuration("10m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Activate().Return(nil)
			mockNow = todayAt(6, 30, 0)
			afterCh <- mockNow
			<-afterNotify

			expectedDuration, _ = time.ParseDuration("1h15m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Deactivate().Return(nil)
			mockNow = todayAt(7, 45, 0)
			afterCh <- mockNow
			<-afterNotify

			expectedDuration, _ = time.ParseDuration("9h48m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Activate().Return(nil)
			mockNow = todayAt(17, 33, 0)
			afterCh <- mockNow
			<-afterNotify
		})

		PIt("should wrap around at the end of the day", func() {
			mockNow = todayAt(20, 04, 23)

			timer.Start()
			<-afterNotify

			expectedDuration, _ := time.ParseDuration("7m37s")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Deactivate().Return(nil)
			mockNow = todayAt(21, 12, 0)
			afterCh <- mockNow
			<-afterNotify

			expectedDuration, _ = time.ParseDuration("9h18m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Activate().Return(nil)
			mockNow = todayAt(6, 30, 0)
			afterCh <- mockNow
			<-afterNotify
		})

		PIt("should handle events added in a non-sequential order", func() {
		})

		PIt("should handle events added after the timer has been started", func() {
		})

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
		})
	})
})

func todayAt(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local)
}
