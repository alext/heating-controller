package scheduler

import (
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
})
