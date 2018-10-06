package units_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/units"
)

var _ = Describe("TimeOfDay", func() {

	DescribeTable("constructing and forammting",
		func(t units.TimeOfDay, expected string) {
			Expect(t.String()).To(Equal(expected))
		},
		Entry("formats the time correctly", units.NewTimeOfDay(12, 34), "12:34"),
		Entry("always uses 2 digits for minutes", units.NewTimeOfDay(12, 5), "12:05"),
		Entry("omits leading zeros for hour", units.NewTimeOfDay(2, 15), "2:15"),
		Entry("supports second precision", units.NewTimeOfDay(12, 34, 35), "12:34:35"),
		Entry("omits seconds if zero", units.NewTimeOfDay(12, 34, 0), "12:34"),
		Entry("always uses 2 digits for seconds", units.NewTimeOfDay(12, 34, 06), "12:34:06"),
	)

	DescribeTable("validity",
		func(t units.TimeOfDay, expected bool) {
			Expect(t.Valid()).To(Equal(expected))
		},
		Entry("a time in the day", units.NewTimeOfDay(12, 34), true),
		Entry("midnight", units.NewTimeOfDay(0, 0, 0), true),
		Entry("one minute before midnight", units.NewTimeOfDay(23, 59), true),
		Entry("one second before midnight", units.NewTimeOfDay(23, 59, 59), true),
		Entry("an invalid time", units.NewTimeOfDay(24, 0, 0), false),
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
		It("handles a second precision TimeOfDay", func() {
			tod = units.NewTimeOfDay(14, 15, 16)
			actual := tod.NextOccuranceAfter(time.Date(2018, 9, 25, 11, 0, 0, 0, london))
			Expect(actual).To(Equal(time.Date(2018, 9, 25, 14, 15, 16, 0, london)))
		})
	})

	DescribeTable("Text marshalling/unmarshalling",
		func(t units.TimeOfDay, serialised string) {
			str := `"` + serialised + `"`
			Expect(json.Marshal(t)).To(BeEquivalentTo(str))

			var actual units.TimeOfDay
			err := json.Unmarshal([]byte(str), &actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(t))
		},
		Entry("encodes and decodes the time correctly", units.NewTimeOfDay(12, 34), "12:34"),
		Entry("handles using 2 digits for minutes", units.NewTimeOfDay(12, 5), "12:05"),
		Entry("omits leading zeros for hour", units.NewTimeOfDay(2, 15), "2:15"),
		Entry("supports second precision", units.NewTimeOfDay(12, 34, 35), "12:34:35"),
		Entry("omits seconds if zero", units.NewTimeOfDay(12, 34, 0), "12:34"),
		Entry("handles using 2 digits for seconds", units.NewTimeOfDay(12, 34, 06), "12:34:06"),
	)

	DescribeTable("UnmarshalText error handling",
		func(input string) {
			var t units.TimeOfDay
			err := json.Unmarshal([]byte(`"`+input+`"`), &t)
			Expect(err).To(HaveOccurred())
		},
		Entry("not given hour and minute", "14"),
		Entry("more than 3 time components", "14:15:16:17"),
		Entry("non-numeric hour", "foo:23"),
		Entry("non-numeric min", "12:foo"),
		Entry("non-numeric second", "12:23:foo"),
	)
})
