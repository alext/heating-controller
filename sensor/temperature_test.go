package sensor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/sensor"
)

var _ = Describe("Temperature", func() {

	DescribeTable("formatting as a string",
		func(input int, expected string) {
			temp := sensor.Temperature(input)
			Expect(temp.String()).To(Equal(expected))
		},
		Entry("returns temp in °C", 19500, "19.5°C"),
		Entry("omits decimals when not needed", 20000, "20°C"),
		Entry("uses decimal places as needed", 19873, "19.873°C"),
		Entry("handles negative values correctly", -5040, "-5.04°C"),
	)
})
