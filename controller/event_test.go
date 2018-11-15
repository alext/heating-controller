package controller_test

import (
	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/units"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {
	Describe("validation", func() {
		It("should be valid for the zero value", func() {
			Expect(controller.Event{}.Valid()).To(BeTrue())
		})

		It("should be invalid with an invalid time", func() {
			Expect(controller.Event{Time: units.NewTimeOfDay(25, 0)}.Valid()).To(BeFalse())
		})
	})
})
