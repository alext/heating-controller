package scheduler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job", func() {
	Describe("validation", func() {
		It("should be valid for the zero value", func() {
			Expect(Job{}.Valid()).To(BeTrue())
		})

		PIt("should be invalid for a negative hour", func() {
			//Expect(Job{Hour: -1}.Valid()).To(BeFalse())
		})

		PIt("should be invalid for an hour greater than 23", func() {
			//Expect(Job{Hour: 24}.Valid()).To(BeFalse())
		})

		PIt("should be invalid for a negative minute", func() {
			//Expect(Job{Min: -1}.Valid()).To(BeFalse())
		})

		PIt("should be invalid for an minute greater than 59", func() {
			//Expect(Job{Min: 60}.Valid()).To(BeFalse())
		})
	})
})
