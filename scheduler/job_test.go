package scheduler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/units"
)

var _ = Describe("Job", func() {
	Describe("validation", func() {
		It("should be valid for the zero value", func() {
			Expect(Job{}.Valid()).To(BeTrue())
		})

		It("should be invalid for a job with an invalid time", func() {
			Expect(Job{Time: units.NewTimeOfDay(25, 0)}.Valid()).To(BeFalse())
		})
	})
})
