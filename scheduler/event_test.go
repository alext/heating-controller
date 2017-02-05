package scheduler

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {
	Describe("validation", func() {
		It("should be valid for the zero value", func() {
			Expect(Event{}.Valid()).To(BeTrue())
		})

		It("should be invalid for a negative hour", func() {
			Expect(Event{Hour: -1}.Valid()).To(BeFalse())
		})

		It("should be invalid for an hour greater than 23", func() {
			Expect(Event{Hour: 24}.Valid()).To(BeFalse())
		})

		It("should be invalid for a negative minute", func() {
			Expect(Event{Min: -1}.Valid()).To(BeFalse())
		})

		It("should be invalid for an minute greater than 59", func() {
			Expect(Event{Min: 60}.Valid()).To(BeFalse())
		})
	})

	Describe("Applying to specific days only", func() {
		var e Event

		It("applies to all days by default", func() {
			for i := 0; i < 7; i++ {
				day := time.Weekday(i)
				Expect(e.ActiveOn(day)).To(BeTrue(),
					"Expected event to be active on "+day.String())
			}
		})

		It("is only active on the specified days", func() {
			e.Days = Monday | Thursday

			Expect(e.ActiveOn(time.Monday)).To(BeTrue(),
				"Expected event to be active on Monday")
			Expect(e.ActiveOn(time.Thursday)).To(BeTrue(),
				"Expected event to be active on Thursday")

			Expect(e.ActiveOn(time.Sunday)).To(BeFalse(),
				"Expected event not to be active on Sunday")
			Expect(e.ActiveOn(time.Tuesday)).To(BeFalse(),
				"Expected event not to be active on Tuesday")
		})
	})
})
