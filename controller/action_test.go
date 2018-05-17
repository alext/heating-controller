package controller_test

import (
	"encoding/json"

	"github.com/alext/heating-controller/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Action", func() {

	Describe("JSON marshalling", func() {
		It("should marshal On", func() {
			Expect(json.Marshal(controller.On)).To(BeEquivalentTo(`"On"`))
		})

		It("should marshal Off", func() {
			Expect(json.Marshal(controller.Off)).To(BeEquivalentTo(`"Off"`))
		})
	})

	Describe("JSON unmarshalling", func() {
		It("should unmarshal On", func() {
			var a controller.Action
			err := json.Unmarshal([]byte(`"On"`), &a)
			Expect(err).NotTo(HaveOccurred())
			Expect(a).To(Equal(controller.On))
		})

		It("should unmarshal Off", func() {
			var a controller.Action
			err := json.Unmarshal([]byte(`"Off"`), &a)
			Expect(err).NotTo(HaveOccurred())
			Expect(a).To(Equal(controller.Off))
		})

		It("should error for an unrecognised string", func() {
			var a controller.Action
			err := json.Unmarshal([]byte(`"Foo"`), &a)
			Expect(err).To(HaveOccurred())
		})
	})
})
