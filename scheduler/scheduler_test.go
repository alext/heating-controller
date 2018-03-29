package scheduler

import (
	"io/ioutil"
	"log"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestScheduler(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)

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

type statefulThing struct {
	mu    sync.Mutex
	state bool
}

func (s *statefulThing) TurnOn()     { s.mu.Lock(); defer s.mu.Unlock(); s.state = true }
func (s *statefulThing) TurnOff()    { s.mu.Lock(); defer s.mu.Unlock(); s.state = false }
func (s *statefulThing) State() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.state }

func (s *statefulThing) ExpectState(st bool) {
	// State should change to st
	EventuallyWithOffset(1, s.State, 100*time.Millisecond, time.Millisecond).Should(Equal(st))
	// and remain there
	ConsistentlyWithOffset(1, s.State, 10*time.Millisecond, time.Millisecond).Should(Equal(st))
}

var _ = Describe("a basic scheduler", func() {
	var (
		mockNow      time.Time
		nowCount     int
		theScheduler Scheduler
		thing        *statefulThing
	)

	BeforeEach(func() {
		timerCh = make(chan time.Time, 1)
		waitNotify = make(chan bool, 1)

		newTimer = func(d time.Duration) timer {
			return dummyTimer{}
		}

		thing = &statefulThing{}
		theScheduler = New("something")

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
					theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: thing.TurnOn})
					theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: thing.TurnOff})
					theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: thing.TurnOn})
					theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: thing.TurnOff})
				})

				It("should apply the previous entry's state on starting", func() {
					mockNow = todayAt(6, 45, 0)

					theScheduler.Start()
					<-waitNotify
					thing.ExpectState(true)
					theScheduler.Stop()

					mockNow = todayAt(12, 00, 0)
					theScheduler.Start()
					<-waitNotify
					thing.ExpectState(false)
				})

				It("should use the last entry from the previous day if necessary", func() {
					mockNow = todayAt(4, 45, 0)

					theScheduler.Start()
					<-waitNotify
					thing.ExpectState(false)
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
			theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: thing.TurnOn})
			theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: thing.TurnOff})
			theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: thing.TurnOn})
			theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: thing.TurnOff})
		})

		It("should fire the given events in order", func() {
			mockNow = todayAt(6, 20, 0)

			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("10m0s"))

			mockNow = todayAt(6, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("1h15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("9h48m0s"))

			mockNow = todayAt(17, 33, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)
		})

		It("should wrap around at the end of the day", func() {
			mockNow = todayAt(20, 04, 23)

			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("1h7m37s"))

			mockNow = todayAt(21, 12, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			nextAt := todayAt(6, 30, 0).AddDate(0, 0, 1)

			Expect(resetParam.String()).To(Equal(nextAt.Sub(mockNow).String()))

			mockNow = nextAt
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)
		})

		It("should handle events added in a non-sequential order", func() {
			theScheduler.AddEvent(Event{Hour: 13, Min: 00, Action: thing.TurnOff})
			theScheduler.AddEvent(Event{Hour: 11, Min: 30, Action: thing.TurnOn})

			mockNow = todayAt(7, 30, 0)

			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("3h45m0s"))

			mockNow = todayAt(11, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("1h30m0s"))

			mockNow = todayAt(13, 0, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("4h33m0s"))
		})

		It("should handle events added after the scheduler has been started", func() {
			mockNow = todayAt(7, 30, 0)

			theScheduler.Start()
			<-waitNotify
			thing.ExpectState(true)

			Expect(resetParam.String()).To(Equal("15m0s"))

			mockNow = todayAt(7, 45, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(false)

			Expect(resetParam.String()).To(Equal("9h48m0s"))

			mockNow = todayAt(9, 30, 0)
			theScheduler.AddEvent(Event{Hour: 11, Min: 30, Action: thing.TurnOn})
			<-waitNotify

			Expect(resetParam.String()).To(Equal("2h0m0s"))

			mockNow = todayAt(11, 30, 0)
			timerCh <- mockNow
			<-waitNotify
			thing.ExpectState(true)
		})
	})

	It("should return an error when adding an invalid event", func() {
		err := theScheduler.AddEvent(Event{Min: -1})
		Expect(err).To(MatchError(ErrInvalidEvent))
		Expect(theScheduler.ReadEvents()).To(HaveLen(0))
	})

	Describe("querying the next event", func() {

		It("should return nil with no events", func() {
			Expect(theScheduler.NextEvent()).To(BeNil())
		})

		Context("with some events", func() {
			BeforeEach(func() {
				theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: thing.TurnOn, Label: "alpha"})
				theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: thing.TurnOn, Label: "bravo"})
				theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: thing.TurnOff, Label: "charlie"})
				theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: thing.TurnOff, Label: "delta"})
			})

			It("should return the next event", func() {
				mockNow = todayAt(6, 0, 0)

				Expect(theScheduler.NextEvent().Label).To(Equal("alpha"))

				mockNow = todayAt(7, 30, 0)

				Expect(theScheduler.NextEvent().Label).To(Equal("charlie"))
			})

			It("should handle the wrap around at the end of the day", func() {
				mockNow = todayAt(21, 30, 0)

				Expect(theScheduler.NextEvent().Label).To(Equal("alpha"))
			})

			Context("with a running timer", func() {
				BeforeEach(func() {
					mockNow = todayAt(14, 0, 0)
					theScheduler.Start()
					<-waitNotify
				})

				It("should return the next event", func() {
					Expect(theScheduler.NextEvent().Label).To(Equal("bravo"))
				})

				It("should return the temporary boost end event when boosted", func() {
					theScheduler.Boost(30*time.Minute, thing.TurnOn)
					<-waitNotify

					e := theScheduler.NextEvent()
					Expect(e.Hour).To(Equal(14))
					Expect(e.Min).To(Equal(30))
				})
			})
		})
	})

	Describe("reading the current schedule", func() {

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
				theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: thing.TurnOn, Label: "alpha"})
				theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: thing.TurnOn, Label: "bravo"})
				theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: thing.TurnOff, Label: "charlie"})
				theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: thing.TurnOff, Label: "delta"})
			})

			It("should return the current event list", func() {
				events := theScheduler.ReadEvents()

				Expect(events).To(HaveLen(4))
				Expect(events[0].Label).To(Equal("alpha"))
				Expect(events[3].Label).To(Equal("delta"))
			})

			It("should return the current event list from a running scheduler", func() {
				theScheduler.Start()
				<-waitNotify

				events := theScheduler.ReadEvents()

				Expect(events).To(HaveLen(4))
				Expect(events[0].Label).To(Equal("alpha"))
				Expect(events[3].Label).To(Equal("delta"))
			})
		})

	})

	Describe("removing an event", func() {
		BeforeEach(func() {
			theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: thing.TurnOn, Label: "alpha"})
			theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: thing.TurnOn, Label: "bravo"})
			theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: thing.TurnOff, Label: "charlie"})
			theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: thing.TurnOff, Label: "delta"})
		})

		Context("with a stopped timer", func() {
			It("should remove the corresponding event from the list", func() {
				theScheduler.RemoveEvent(Event{Hour: 7, Min: 45, Action: thing.TurnOff, Label: "charlie"})

				Expect(theScheduler.ReadEvents()).To(HaveLen(3))
			})

			It("should do nothing if the event isn't in the scheduler", func() {
				theScheduler.RemoveEvent(Event{Hour: 7, Min: 50, Action: thing.TurnOn, Label: "foo"})

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
				theScheduler.RemoveEvent(Event{Hour: 7, Min: 45, Action: thing.TurnOff, Label: "charlie"})

				Expect(theScheduler.ReadEvents()).To(HaveLen(3))
			})

			It("should do nothing if the event isn't in the scheduler", func() {
				theScheduler.RemoveEvent(Event{Hour: 7, Min: 50, Action: thing.TurnOn, Label: "foo"})

				Expect(theScheduler.ReadEvents()).To(HaveLen(4))
			})

			It("should reschedule if the removed event was the next event", func() {
				mockNow = todayAt(15, 0, 0)
				theScheduler.RemoveEvent(Event{Hour: 17, Min: 33, Action: thing.TurnOn, Label: "bravo"})

				<-waitNotify
				Expect(resetParam.String()).To(Equal("6h12m0s"))
			})
		})

	})

	Describe("boost function", func() {

		Context("a scheduler with events", func() {
			BeforeEach(func() {
				theScheduler.AddEvent(Event{Hour: 6, Min: 30, Action: thing.TurnOn, Label: "alpha"})
				theScheduler.AddEvent(Event{Hour: 7, Min: 45, Action: thing.TurnOff, Label: "bravo"})
				theScheduler.AddEvent(Event{Hour: 17, Min: 33, Action: thing.TurnOn, Label: "charlie"})
				theScheduler.AddEvent(Event{Hour: 21, Min: 12, Action: thing.TurnOff, Label: "delta"})
			})

			It("should activate the output for the specified duration, then resume the schedule", func() {
				mockNow = todayAt(14, 0, 0)
				theScheduler.Start()

				<-waitNotify
				thing.ExpectState(false)

				mockNow = todayAt(14, 30, 0)
				theScheduler.Boost(40*time.Minute, thing.TurnOn)

				<-waitNotify
				thing.ExpectState(true)
				Expect(resetParam.String()).To(Equal("40m0s"))
				Expect(theScheduler.Boosted()).To(BeTrue())

				mockNow = todayAt(15, 10, 0)
				timerCh <- mockNow
				<-waitNotify
				thing.ExpectState(false)
				Expect(resetParam.String()).To(Equal("2h23m0s"))
				Expect(theScheduler.Boosted()).To(BeFalse())
			})

			It("should allow cancelling the boost", func() {
				mockNow = todayAt(14, 0, 0)
				theScheduler.Start()
				<-waitNotify
				thing.ExpectState(false)

				mockNow = todayAt(14, 30, 0)
				theScheduler.Boost(40*time.Minute, thing.TurnOn)
				<-waitNotify
				thing.ExpectState(true)

				mockNow = todayAt(14, 55, 0)
				theScheduler.CancelBoost()
				<-waitNotify

				Expect(theScheduler.Boosted()).To(BeFalse())
				Expect(resetParam.String()).To(Equal("2h38m0s"))
				thing.ExpectState(false)
			})

			Context("boosting until the next event", func() {
				It("activates the output, then resumes the schedule", func() {
					mockNow = todayAt(14, 0, 0)
					theScheduler.Start()

					<-waitNotify
					thing.ExpectState(false)

					mockNow = todayAt(14, 30, 0)
					theScheduler.Boost(0, thing.TurnOn)

					<-waitNotify
					thing.ExpectState(true)
					Expect(theScheduler.Boosted()).To(BeTrue())
					Expect(theScheduler.NextEvent().Label).To(Equal("charlie"))

					mockNow = todayAt(17, 33, 0)
					timerCh <- mockNow
					<-waitNotify
					thing.ExpectState(true)
					Expect(resetParam.String()).To(Equal("3h39m0s"))
					Expect(theScheduler.Boosted()).To(BeFalse())
				})
			})

			Context("overlapping an upcoming event", func() {
				BeforeEach(func() {
					mockNow = todayAt(16, 0, 0)
					theScheduler.Start()
					<-waitNotify
					thing.ExpectState(false)
				})

				It("should overlap an upcoming event, returning to the overlapped event state after the boost", func() {
					mockNow = todayAt(17, 00, 0)
					theScheduler.Boost(40*time.Minute, thing.TurnOn)

					<-waitNotify
					thing.ExpectState(true)
					Expect(theScheduler.Boosted()).To(BeTrue())

					mockNow = todayAt(17, 40, 0)
					timerCh <- mockNow
					<-waitNotify
					thing.ExpectState(true)
					Expect(resetParam.String()).To(Equal("3h32m0s"))
					Expect(theScheduler.Boosted()).To(BeFalse())
				})

				It("cancelling the boost should retain the state of the overlapped event", func() {
					mockNow = todayAt(17, 00, 0)
					theScheduler.Boost(40*time.Minute, thing.TurnOn)
					<-waitNotify

					mockNow = todayAt(17, 35, 0)
					theScheduler.CancelBoost()
					<-waitNotify

					thing.ExpectState(true)
					Expect(theScheduler.Boosted()).To(BeFalse())
					Expect(resetParam.String()).To(Equal("3h37m0s"))
				})
			})
		})

		Context("a scheduler with no events", func() {
			BeforeEach(func() {
				mockNow = todayAt(6, 0, 0)
				theScheduler.Start()
				<-waitNotify
			})

			It("should activate the output for the specified duraton, then return to sleeping for 24 hrs", func() {

				mockNow = todayAt(7, 30, 0)
				theScheduler.Boost(45*time.Minute, thing.TurnOn)

				<-waitNotify
				thing.ExpectState(true)
				Expect(resetParam.String()).To(Equal("45m0s"))
				Expect(theScheduler.Boosted()).To(BeTrue())
				<-waitNotify // Boosted() triggers another loop

				mockNow = todayAt(8, 15, 0)
				timerCh <- mockNow
				<-waitNotify
				Expect(resetParam.String()).To(Equal("24h0m0s"))
				Expect(theScheduler.Boosted()).To(BeFalse())
			})

			It("should allow cancelling the boost", func() {
				theScheduler.Boost(45*time.Minute, thing.TurnOn)
				<-waitNotify
				thing.ExpectState(true)

				mockNow = todayAt(6, 26, 0)
				theScheduler.CancelBoost()
				<-waitNotify

				Expect(theScheduler.Boosted()).To(BeFalse())
				Expect(resetParam.String()).To(Equal("24h0m0s"))
			})

			Context("attempting to boost until the next event", func() {
				It("should boost for an hour", func() {
					mockNow = todayAt(7, 30, 0)
					theScheduler.Boost(0, thing.TurnOn)

					<-waitNotify
					thing.ExpectState(true)
					Expect(resetParam.String()).To(Equal("1h0m0s"))
					Expect(theScheduler.Boosted()).To(BeTrue())
				})
			})
		})

	})
})

func todayAt(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, time.Local)
}
