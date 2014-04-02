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
		theTimer    Timer
		afterParam  time.Duration
		afterCh     chan time.Time
		afterNotify chan bool
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		output = mock_output.NewMockOutput(mockCtrl)
		theTimer = New(output)

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
		theTimer.Stop()
		mockCtrl.Finish()
	})

	Describe("starting and stopping the timer", func() {
		It("should not be running when newly created", func() {
			Expect(theTimer.Running()).To(BeFalse())
		})

		It("should start the timer", func() {
			theTimer.Start()
			Expect(theTimer.Running()).To(BeTrue())
		})

		It("should do nothing when attempting to start a running timer", func() {
			theTimer.Start()
			theTimer.Start()
			<-afterNotify

			Expect(nowCount).To(Equal(1))
		})

		It("should stop the timer", func() {
			theTimer.Start()
			theTimer.Stop()
			Expect(theTimer.Running()).To(BeFalse())
		})

		It("should do nothing when attempting to stop a non-running timer", func(done Done) {
			theTimer.Stop()
			close(done)
		}, 0.5)
	})

	It("should continuously sleep for a day when started with no entries", func() {
		mockNow = todayAt(6, 20, 0)

		theTimer.Start()
		<-afterNotify

		expectedDuration, _ := time.ParseDuration("24h")
		Expect(afterParam).To(Equal(expectedDuration))

		mockNow = mockNow.AddDate(0, 0, 1)
		afterCh <- mockNow
		<-afterNotify

		expectedDuration, _ = time.ParseDuration("24h")
		Expect(afterParam).To(Equal(expectedDuration))
	})

	Describe("firing events as scheduled", func() {

		BeforeEach(func() {
			theTimer.AddEntry(6, 30, TurnOn)
			theTimer.AddEntry(7, 45, TurnOff)
			theTimer.AddEntry(17, 33, TurnOn)
			theTimer.AddEntry(21, 12, TurnOff)
		})

		It("should fire the given events in order", func() {
			mockNow = todayAt(6, 20, 0)

			theTimer.Start()
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

		It("should wrap around at the end of the day", func() {
			mockNow = todayAt(20, 04, 23)

			theTimer.Start()
			<-afterNotify

			expectedDuration, _ := time.ParseDuration("1h7m37s")
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

		It("should handle events added in a non-sequential order", func() {
			theTimer.AddEntry(13, 00, TurnOff)
			theTimer.AddEntry(11, 30, TurnOn)

			mockNow = todayAt(7, 30, 0)

			theTimer.Start()
			<-afterNotify

			expectedDuration, _ := time.ParseDuration("15m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Deactivate().Return(nil)
			mockNow = todayAt(7, 45, 0)
			afterCh <- mockNow
			<-afterNotify

			expectedDuration, _ = time.ParseDuration("3h45m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Activate().Return(nil)
			mockNow = todayAt(11, 30, 0)
			afterCh <- mockNow
			<-afterNotify

			expectedDuration, _ = time.ParseDuration("1h30m")
			Expect(afterParam).To(Equal(expectedDuration))

			output.EXPECT().Deactivate().Return(nil)
			mockNow = todayAt(13, 0, 0)
			afterCh <- mockNow
			<-afterNotify

			expectedDuration, _ = time.ParseDuration("4h33m")
			Expect(afterParam).To(Equal(expectedDuration))
		})

		PIt("should handle events added after the timer has been started", func() {
		})
	})
})

func todayAt(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, time.Local)
}
