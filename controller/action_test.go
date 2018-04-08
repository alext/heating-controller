package controller_test

import (
	"encoding/json"

	"github.com/alext/heating-controller/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Action", func() {

	Describe("JSON marshalling", func() {
		It("should marshal TurnOn", func() {
			Expect(json.Marshal(controller.TurnOn)).To(BeEquivalentTo(`"On"`))
		})

		It("should marshal TurnOff", func() {
			Expect(json.Marshal(controller.TurnOff)).To(BeEquivalentTo(`"Off"`))
		})
	})

	Describe("JSON unmarshalling", func() {
		It("should unmarshal TurnOn", func() {
			var a controller.Action
			err := json.Unmarshal([]byte(`"On"`), &a)
			Expect(err).NotTo(HaveOccurred())
			Expect(a).To(Equal(controller.TurnOn))
		})

		It("should unmarshal TurnOn", func() {
			var a controller.Action
			err := json.Unmarshal([]byte(`"Off"`), &a)
			Expect(err).NotTo(HaveOccurred())
			Expect(a).To(Equal(controller.TurnOff))
		})

		It("should error for an unrecognised string", func() {
			var a controller.Action
			err := json.Unmarshal([]byte(`"Foo"`), &a)
			Expect(err).To(HaveOccurred())
		})
	})
})
