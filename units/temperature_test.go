package units_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/units"
)

var _ = Describe("Temperature", func() {

	DescribeTable("formatting as a string",
		func(input int, expected string) {
			temp := units.Temperature(input)
			Expect(temp.String()).To(Equal(expected))
		},
		Entry("returns temp in °C", 19500, "19.5°C"),
		Entry("omits decimals when not needed", 20000, "20°C"),
		Entry("uses decimal places as needed", 19873, "19.873°C"),
		Entry("handles negative values correctly", -5040, "-5.04°C"),
	)

	DescribeTable("parsing a string",
		func(input string, expected int, expectValid bool) {
			expectedTemp := units.Temperature(expected)
			actual, err := units.ParseTemperature(input)
			if expectValid {
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expectedTemp))
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("handles integers", "19", 19000, true),
		Entry("handles decimals", "19.5", 19500, true),
		Entry("handles negatives", "-3", -3000, true),
		Entry("handles integer with unit", "20°C", 20000, true),
		Entry("handles decimal with unit", "20.1°C", 20100, true),
		Entry("errors with invalid input", "foo", 0, false),
	)
})
