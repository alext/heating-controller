package units_test

import (
	"time"

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

	Describe("NextOccuranceAfter", func() {
		var (
			tod    units.TimeOfDay
			london *time.Location
		)

		BeforeEach(func() {
			var err error
			tod = units.NewTimeOfDay(14, 15)
			london, err = time.LoadLocation("Europe/London")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the time.Time of the next occurance after the given time", func() {
			actual := tod.NextOccuranceAfter(time.Date(2018, 9, 25, 11, 0, 0, 0, london))
			Expect(actual).To(Equal(time.Date(2018, 9, 25, 14, 15, 0, 0, london)))
		})
		It("wraps around to the following day if current is after TOD", func() {
			actual := tod.NextOccuranceAfter(time.Date(2018, 9, 25, 15, 0, 0, 0, london))
			Expect(actual).To(Equal(time.Date(2018, 9, 26, 14, 15, 0, 0, london)))
		})
		It("returns a time in the same timezone as the input", func() {
			actual := tod.NextOccuranceAfter(time.Date(2018, 9, 25, 15, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60)))
			Expect(actual).To(Equal(time.Date(2018, 9, 26, 14, 15, 0, 0, time.FixedZone("UTC+3", 3*60*60))))
		})
		It("returns the specified time correctly on DST boundaries", func() {
			actual := tod.NextOccuranceAfter(time.Date(2018, 10, 27, 18, 0, 0, 0, london))
			Expect(actual).To(Equal(time.Date(2018, 10, 28, 14, 15, 0, 0, london)))
		})
	})
})
