package scheduler

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

	RunSpecs(t, "Scheduler")
}

var (
	timerCh    chan time.Time
	resetParam time.Duration
	waitNotify chan bool
)

type dummyTimer struct{}

func (tmr dummyTimer) Channel() <-chan time.Time {
	// Notify the channel, but don't block if nothing is listening.
	select {
	case waitNotify <- true:
	default:
	}

	return timerCh
}

func (tmr dummyTimer) Reset(d time.Duration) bool {
	resetParam = d
	return true
}

func (tmr dummyTimer) Stop() bool {
	return true
}

var _ = Describe("a basic scheduler", func() {
	var (
		theOutput    output.Output
		mockNow      time.Time
		nowCount     int
		theScheduler Scheduler
	)

	BeforeEach(func() {
		timerCh = make(chan time.Time, 1)
		waitNotify = make(chan bool, 1)

		newTimer = func(d time.Duration) timer {
			return dummyTimer{}
		}

		theOutput = output.Virtual("out")
		theScheduler = New(theOutput)

		mockNow = time.Now()
		nowCount = 0
		time_Now = func() time.Time {
			nowCount++
			return mockNow
		}
	})

	AfterEach(func() {
		theScheduler.Stop()
	})

	Describe("starting and stopping the scheduler", func() {
		It("should not be running when newly created", func() {
			Expect(theScheduler.Running()).To(BeFalse())
		})

		It("should start the scheduler", func() {
			theScheduler.Start()
			Expect(theScheduler.Running()).To(BeTrue())
		})

		It("should do nothing when attempting to start a running scheduler", func() {
			theScheduler.Start()
			theScheduler.Start()
			<-waitNotify

			Expect(nowCount).To(Equal(1))
		})

		It("should stop the scheduler", func() {
			theScheduler.Start()
			theScheduler.Stop()
			Expect(theScheduler.Running()).To(BeFalse())
		})

		It("should do nothing when attempting to stop a non-running scheduler", func(done Done) {
			theScheduler.Stop()
			close(done)
		}, 0.5)

		Describe("setting the initial output state", func() {
			Context("with some entries", func() {
				BeforeEach(func() {
					theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: TurnOn})
					theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: TurnOff})
					theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: TurnOn})
					theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: TurnOff})
				})

				It("should apply the previous entry's state on starting", func() {
					mockNow = todayAt(6, 45, 0)

					theScheduler.Start()
					<-waitNotify
					Expect(theOutput.Active()).To(BeTrue())
					theScheduler.Stop()

					mockNow = todayAt(12, 00, 0)
					theScheduler.Start()
					<-waitNotify
					Expect(theOutput.Active()).To(BeFalse())
				})

				It("should use the last entry from the previous day if necessary", func() {
					mockNow = todayAt(4, 45, 0)

					theScheduler.Start()
					<-waitNotify
					Expect(theOutput.Active()).To(BeFalse())
				})
			})

			It("should do nothing with no entries", func() {
				theScheduler.Start()
				<-waitNotify
				// expect it not to blow up
			})
		})
	})

	It("should continuously sleep for a day when started with no entries", func() {
		mockNow = todayAt(6, 20, 0)

		theScheduler.Start()
		<-waitNotify

		Expect(resetParam.String()).To(Equal("24h0m0s"))

		mockNow = mockNow.AddDate(0, 0, 1)
		timerCh <- mockNow
		<-waitNotify

		Expect(resetParam.String()).To(Equal("24h0m0s"))
	})

	Describe("firing events as scheduled", func() {

		BeforeEach(func() {
			theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: TurnOn})
			theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: TurnOff})
			theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: TurnOn})
			theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: TurnOff})
		})

		It("should fire the given events in order", func() {
			mockNow = todayAt(6, 20, 0)

			theScheduler.Start()
			<-waitNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(resetParam.String()).To(Equal("10m0s"))

			mockNow = todayAt(6, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(resetParam.String()).To(Equal("1h15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(resetParam.String()).To(Equal("9h48m0s"))

			mockNow = todayAt(17, 33, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeTrue())
		})

		It("should wrap around at the end of the day", func() {
			mockNow = todayAt(20, 04, 23)

			theScheduler.Start()
			<-waitNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(resetParam.String()).To(Equal("1h7m37s"))

			mockNow = todayAt(21, 12, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(resetParam.String()).To(Equal("9h18m0s"))

			mockNow = todayAt(6, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeTrue())
		})

		It("should handle events added in a non-sequential order", func() {
			theScheduler.AddEvent(Event{Hour: 13, Min: 00, Action: TurnOff})
			theScheduler.AddEvent(Event{Hour: 11, Min: 30, Action: TurnOn})

			mockNow = todayAt(7, 30, 0)

			theScheduler.Start()
			<-waitNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(resetParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(resetParam.String()).To(Equal("3h45m0s"))

			mockNow = todayAt(11, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(resetParam.String()).To(Equal("1h30m0s"))

			mockNow = todayAt(13, 0, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(resetParam.String()).To(Equal("4h33m0s"))
		})

		It("should handle events added after the scheduler has been started", func() {
			mockNow = todayAt(7, 30, 0)

			theScheduler.Start()
			<-waitNotify
			Expect(theOutput.Active()).To(BeTrue())

			Expect(resetParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeFalse())

			Expect(resetParam.String()).To(Equal("9h48m0s"))

			mockNow = todayAt(9, 30, 0)
			theScheduler.AddEvent(Event{Hour: 11, Min: 30, Action: TurnOn})
			<-waitNotify

			Expect(resetParam.String()).To(Equal("2h0m0s"))

			mockNow = todayAt(11, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			Expect(theOutput.Active()).To(BeTrue())
		})
	})

	Describe("querying the next event", func() {

		It("should return nil with no events", func() {
			Expect(theScheduler.NextEvent()).To(BeNil())
		})

		Context("with some events", func() {
			BeforeEach(func() {
				theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: TurnOn})
				theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: TurnOn})
				theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: TurnOff})
				theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: TurnOff})
			})

			It("should return the next event", func() {
				mockNow = todayAt(6, 0, 0)

				Expect(theScheduler.NextEvent()).To(Equal(&Event{Hour: 6, Min: 30, Action: TurnOn}))

				mockNow = todayAt(7, 30, 0)

				Expect(theScheduler.NextEvent()).To(Equal(&Event{Hour: 7, Min: 45, Action: TurnOff}))
			})

			It("should handle the wrap around at the end of the day", func() {
				mockNow = todayAt(21, 30, 0)

				Expect(theScheduler.NextEvent()).To(Equal(&Event{Hour: 6, Min: 30, Action: TurnOn}))
			})

			Context("with a running timer", func() {
				BeforeEach(func() {
					mockNow = todayAt(14, 0, 0)
					theScheduler.Start()
					<-waitNotify
				})

				It("should return the next event", func() {

					Expect(theScheduler.NextEvent()).To(Equal(&Event{Hour: 17, Min: 33, Action: TurnOn}))
				})

				It("should return the temporary boost end event when boosted", func() {
					theScheduler.Boost(30 * time.Minute)
					<-waitNotify

					Expect(theScheduler.NextEvent()).To(Equal(&Event{Hour: 14, Min: 30, Action: TurnOff}))
				})
			})
		})
	})

	Describe("readling the current schedule", func() {

		It("should return an empty list for a stopped scheduler with no events", func() {
			Expect(theScheduler.ReadEvents()).To(BeEmpty())
		})

		It("should return an empty list for a running scheduler with no events", func() {
			theScheduler.Start()
			<-waitNotify

			Expect(theScheduler.ReadEvents()).To(BeEmpty())
		})

		Context("with some events", func() {
			BeforeEach(func() {
				theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: TurnOn})
				theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: TurnOff})
				theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: TurnOn})
				theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: TurnOff})
			})

			It("should return the current event list", func() {
				events := theScheduler.ReadEvents()

				Expect(events).To(HaveLen(4))
				Expect(events[0]).To(Equal(Event{Hour: 6, Min: 30, Action: TurnOn}))
				Expect(events[3]).To(Equal(Event{Hour: 21, Min: 12, Action: TurnOff}))
			})

			It("should return the current event list from a running scheduler", func() {
				theScheduler.Start()
				<-waitNotify

				events := theScheduler.ReadEvents()

				Expect(events).To(HaveLen(4))
				Expect(events[0]).To(Equal(Event{Hour: 6, Min: 30, Action: TurnOn}))
				Expect(events[3]).To(Equal(Event{Hour: 21, Min: 12, Action: TurnOff}))
			})
		})

	})

	Describe("removing an event", func() {
		BeforeEach(func() {
			theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: TurnOn})
			theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: TurnOff})
			theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: TurnOn})
			theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: TurnOff})
		})

		Context("with a stopped timer", func() {
			It("should remove the corresponding event from the list", func() {
				theScheduler.RemoveEvent(Event{Hour: 7, Min: 45, Action: TurnOff})

				Expect(theScheduler.ReadEvents()).To(HaveLen(3))
			})

			It("should do nothing if the event isn't in the scheduler", func() {
				theScheduler.RemoveEvent(Event{Hour: 7, Min: 45, Action: TurnOn})

				Expect(theScheduler.ReadEvents()).To(HaveLen(4))
			})
		})

		Context("with a running scheduler", func() {
			BeforeEach(func() {
				mockNow = todayAt(14, 0, 0)
				theScheduler.Start()
				<-waitNotify
			})

			It("should remove the event from the list", func() {
				theScheduler.RemoveEvent(Event{Hour: 7, Min: 45, Action: TurnOff})

				Expect(theScheduler.ReadEvents()).To(HaveLen(3))
			})

			It("should do nothing if the event isn't in the scheduler", func() {
				theScheduler.RemoveEvent(Event{Hour: 7, Min: 45, Action: TurnOn})

				Expect(theScheduler.ReadEvents()).To(HaveLen(4))
			})

			It("should reschedule if the removed event was the next event", func() {
				mockNow = todayAt(15, 0, 0)
				theScheduler.RemoveEvent(Event{Hour: 17, Min: 33, Action: TurnOn})

				<-waitNotify
				Expect(resetParam.String()).To(Equal("6h12m0s"))

			})
		})

	})

	Describe("boost function", func() {

		Context("a scheduler with no events", func() {
			BeforeEach(func() {
				mockNow = todayAt(6, 0, 0)
				theScheduler.Start()
				<-waitNotify
			})

			It("should activate the output for the specified duraton", func() {

				mockNow = todayAt(7, 30, 0)
				theScheduler.Boost(45 * time.Minute)

				<-waitNotify
				Expect(theOutput.Active()).To(BeTrue())
				Expect(resetParam.String()).To(Equal("45m0s"))
				Expect(theScheduler.Boosted()).To(BeTrue())

				mockNow = todayAt(8, 15, 0)
				timerCh <- mockNow
				<-waitNotify
				Expect(theOutput.Active()).To(BeFalse())
				Expect(resetParam.String()).To(Equal("24h0m0s"))
				Expect(theScheduler.Boosted()).To(BeFalse())
			})

			It("should allow cancelling the boost", func() {
				theScheduler.Boost(45 * time.Minute)
				<-waitNotify

				mockNow = todayAt(6, 26, 0)
				theScheduler.CancelBoost()
				<-waitNotify

				Expect(theOutput.Active()).To(BeFalse())
				Expect(theScheduler.Boosted()).To(BeFalse())
				Expect(resetParam.String()).To(Equal("24h0m0s"))
			})
		})

		Context("a scheduler with events", func() {
			BeforeEach(func() {
				theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: TurnOn})
				theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: TurnOff})
				theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: TurnOn})
				theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: TurnOff})
			})

			It("should activate the output for the specified duration, then resume the schedule", func() {
				mockNow = todayAt(14, 0, 0)
				theScheduler.Start()

				<-waitNotify

				mockNow = todayAt(14, 30, 0)
				theScheduler.Boost(40 * time.Minute)

				<-waitNotify
				Expect(theOutput.Active()).To(BeTrue())
				Expect(resetParam.String()).To(Equal("40m0s"))
				Expect(theScheduler.Boosted()).To(BeTrue())

				mockNow = todayAt(15, 10, 0)
				timerCh <- mockNow
				<-waitNotify
				Expect(theOutput.Active()).To(BeFalse())
				Expect(resetParam.String()).To(Equal("2h23m0s"))
				Expect(theScheduler.Boosted()).To(BeFalse())
			})

			It("should allow cancelling the boost", func() {
				mockNow = todayAt(14, 0, 0)
				theScheduler.Start()
				<-waitNotify

				mockNow = todayAt(14, 30, 0)
				theScheduler.Boost(40 * time.Minute)
				<-waitNotify

				mockNow = todayAt(14, 55, 0)
				theScheduler.CancelBoost()
				<-waitNotify

				Expect(theOutput.Active()).To(BeFalse())
				Expect(theScheduler.Boosted()).To(BeFalse())
				Expect(resetParam.String()).To(Equal("2h38m0s"))
			})

			Context("overlapping an upcoming TurnOn event", func() {
				BeforeEach(func() {
					mockNow = todayAt(16, 0, 0)
					theScheduler.Start()
					<-waitNotify
				})

				It("should overlap an upcoming TurnOn event", func() {
					mockNow = todayAt(17, 00, 0)
					theScheduler.Boost(40 * time.Minute)

					<-waitNotify
					Expect(theOutput.Active()).To(BeTrue())
					Expect(theScheduler.Boosted()).To(BeTrue())

					mockNow = todayAt(17, 33, 0)
					timerCh <- mockNow
					<-waitNotify
					Expect(theOutput.Active()).To(BeTrue())
					Expect(resetParam.String()).To(Equal("3h39m0s"))
					Expect(theScheduler.Boosted()).To(BeFalse())
				})

				It("cancelling the boost should retain the state of the overlapped event", func() {
					mockNow = todayAt(17, 00, 0)
					theScheduler.Boost(40 * time.Minute)
					<-waitNotify

					mockNow = todayAt(17, 35, 0)
					theScheduler.CancelBoost()
					<-waitNotify

					Expect(theOutput.Active()).To(BeTrue())
					Expect(theScheduler.Boosted()).To(BeFalse())
					Expect(resetParam.String()).To(Equal("3h37m0s"))
				})
			})

			Context("extending beyond the next TurnOff event", func() {
				BeforeEach(func() {
					mockNow = todayAt(7, 25, 0)
					theScheduler.Start()
					<-waitNotify
				})

				It("should extend beyond next TurnOff event", func() {
					mockNow = todayAt(7, 30, 0)
					theScheduler.Boost(30 * time.Minute)

					<-waitNotify
					Expect(theOutput.Active()).To(BeTrue())
					Expect(resetParam.String()).To(Equal("30m0s"))
					Expect(theScheduler.Boosted()).To(BeTrue())

					mockNow = todayAt(8, 0, 0)
					timerCh <- mockNow
					<-waitNotify
					Expect(theOutput.Active()).To(BeFalse())
					Expect(resetParam.String()).To(Equal("9h33m0s"))
					Expect(theScheduler.Boosted()).To(BeFalse())
				})

				It("cancelling the boost should retain the event state", func() {
					mockNow = todayAt(7, 30, 0)
					theScheduler.Boost(30 * time.Minute)
					<-waitNotify

					mockNow = todayAt(7, 40, 0)
					theScheduler.CancelBoost()
					<-waitNotify

					Expect(theOutput.Active()).To(BeTrue())
					Expect(theScheduler.Boosted()).To(BeFalse())
					Expect(resetParam.String()).To(Equal("5m0s"))
				})
			})
		})
	})
})

func todayAt(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, time.Local)
}
