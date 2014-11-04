package timer

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
)

func TestOutput(t *testing.T) {
	RegisterFailHandler(Fail)

	logger.Level = logger.WARN

	RunSpecs(t, "Timer")
}

var _ = Describe("a basic timer", func() {
	var (
		theOutput   output.Output
		mockNow     time.Time
		nowCount    int
		theTimer    Timer
		afterParam  time.Duration
		afterCh     chan time.Time
		afterNotify chan bool
	)

	BeforeEach(func() {
		theOutput = output.Virtual("out")
		theTimer = New(theOutput)

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

		Describe("setting the initial output state", func() {
			Context("with some entries", func() {
				BeforeEach(func() {
					theTimer.AddEntry(6, 30, TurnOn)
					theTimer.AddEntry(7, 45, TurnOff)
					theTimer.AddEntry(17, 33, TurnOn)
					theTimer.AddEntry(21, 12, TurnOff)
				})

				It("should apply the previous entry's state on starting", func() {
					mockNow = todayAt(6, 45, 0)

					theTimer.Start()
					<-afterNotify
					Expect(theOutput.Active()).To(BeTrue())
					theTimer.Stop()

					mockNow = todayAt(12, 00, 0)
					theTimer.Start()
					<-afterNotify
					Expect(theOutput.Active()).To(BeFalse())
				})

				It("should use the last entry from the previous day if necessary", func() {
					mockNow = todayAt(4, 45, 0)

					theTimer.Start()
					<-afterNotify
					Expect(theOutput.Active()).To(BeFalse())
				})
			})

			It("should do nothing with no entries", func() {
				theTimer.Start()
				<-afterNotify
				// expect it not to blow up
			})
		})
	})

	It("should continuously sleep for a day when started with no entries", func() {
		mockNow = todayAt(6, 20, 0)

		theTimer.Start()
		<-afterNotify

		Expect(afterParam.String()).To(Equal("24h0m0s"))

		mockNow = mockNow.AddDate(0, 0, 1)
		afterCh <- mockNow
		<-afterNotify

		Expect(afterParam.String()).To(Equal("24h0m0s"))
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
			Expect(theOutput.Active()).To(BeFalse())

			Expect(afterParam.String()).To(Equal("10m0s"))

			mockNow = todayAt(6, 30, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(afterParam.String()).To(Equal("1h15m0s"))

			mockNow = todayAt(7, 45, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(afterParam.String()).To(Equal("9h48m0s"))

			mockNow = todayAt(17, 33, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeTrue())
		})

		It("should wrap around at the end of the day", func() {
			mockNow = todayAt(20, 04, 23)

			theTimer.Start()
			<-afterNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(afterParam.String()).To(Equal("1h7m37s"))

			mockNow = todayAt(21, 12, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(afterParam.String()).To(Equal("9h18m0s"))

			mockNow = todayAt(6, 30, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeTrue())
		})

		It("should handle events added in a non-sequential order", func() {
			theTimer.AddEntry(13, 00, TurnOff)
			theTimer.AddEntry(11, 30, TurnOn)

			mockNow = todayAt(7, 30, 0)

			theTimer.Start()
			<-afterNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(afterParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(7, 45, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(afterParam.String()).To(Equal("3h45m0s"))

			mockNow = todayAt(11, 30, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(afterParam.String()).To(Equal("1h30m0s"))

			mockNow = todayAt(13, 0, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(afterParam.String()).To(Equal("4h33m0s"))
		})

		It("should handle events added after the timer has been started", func() {
			mockNow = todayAt(7, 30, 0)

			theTimer.Start()
			<-afterNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(afterParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(7, 45, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(afterParam.String()).To(Equal("9h48m0s"))

			mockNow = todayAt(9, 30, 0)
			theTimer.AddEntry(11, 30, TurnOn)
			<-afterNotify

			Expect(afterParam.String()).To(Equal("2h0m0s"))

			mockNow = todayAt(11, 30, 0)
			afterCh <- mockNow
			<-afterNotify
			Expect(theOutput.Active()).To(BeTrue())
		})

		PIt("should handle an entry being added at the moment the next event is due to fire", func() {
			mockNow = todayAt(6, 15, 0)
			theTimer.Start()
			<-afterNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(afterParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(6, 30, 0)

			theTimer.AddEntry(12, 45, TurnOn)
			<-afterNotify

			// Should still be trying to fire the 6:30 entry
			Expect(afterParam.String()).To(Equal("0"))
		})
	})
})

func todayAt(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, time.Local)
}
