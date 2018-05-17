package controller_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
)

var _ = Describe("Action", func() {

	DescribeTable("string representation",
		func(a controller.Action, str string) {
			Expect(a.String()).To(Equal(str))
		},
		Entry("On", controller.On, "On"),
		Entry("Off", controller.Off, "Off"),
		Entry("SetTarget", controller.SetTarget, "SetTarget"),
		Entry("IncreaseTarget", controller.IncreaseTarget, "IncreaseTarget"),
		Entry("DecreaseTarget", controller.DecreaseTarget, "DecreaseTarget"),
	)

	DescribeTable("JSON marshalling/unmarshalling",
		func(a controller.Action) {
			str := `"` + a.String() + `"`
			Expect(json.Marshal(a)).To(BeEquivalentTo(str))

			var actual controller.Action
			err := json.Unmarshal([]byte(str), &actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(a))
		},
		Entry("On", controller.On),
		Entry("Off", controller.On),
		Entry("SetTarget", controller.SetTarget),
		Entry("IncreaseTarget", controller.IncreaseTarget),
		Entry("DecreaseTarget", controller.DecreaseTarget),
	)

	It("JSON unmarshal should error for an unrecognised string", func() {
		var a controller.Action
		err := json.Unmarshal([]byte(`"Foo"`), &a)
		Expect(err).To(HaveOccurred())
	})
})
