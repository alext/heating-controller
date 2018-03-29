package controller_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/output"
)

var _ = Describe("A heating zone", func() {

	Describe("constructing a zone", func() {
		var (
			out output.Output
		)

		BeforeEach(func() {
			out = output.Virtual("anything")
		})

		It("should return a zone with the given id", func() {
			Expect(controller.NewZone("foo", out).ID).To(Equal("foo"))
		})

		It("should construct a scheduler", func() {
			z := controller.NewZone("foo", out)
			Expect(z.Scheduler).NotTo(BeNil())
		})
	})

})
