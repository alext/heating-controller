package units_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/units"
)

var _ = Describe("TimeOfDay", func() {

	DescribeTable("constructing and forammting",
		func(hour, min int, expected string) {
			t := units.NewTimeOfDay(hour, min)
			Expect(t.String()).To(Equal(expected))
		},
		Entry("formats the time correctly", 12, 34, "12:34"),
		Entry("always uses 2 digits for minutes", 12, 5, "12:05"),
		Entry("omits leading zeros for hour", 2, 15, "2:15"),
	)
})
