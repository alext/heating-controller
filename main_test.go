package main

import (
	"fmt"
	"testing"

	"code.google.com/p/gomock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/timer"
	"github.com/alext/heating-controller/timer/mock_timer"
	"github.com/alext/heating-controller/zone"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main")
}

type testZoneAdder struct {
	Zones []*zone.Zone
}

func (t *testZoneAdder) AddZone(z *zone.Zone) {
	t.Zones = append(t.Zones, z)
}

var _ = Describe("Reading zones from cmdline", func() {
	var (
		srv *testZoneAdder
	)

	BeforeEach(func() {
		srv = &testZoneAdder{make([]*zone.Zone, 0)}
		output_New = func(id string, pin int) (output.Output, error) {
			out := output.Virtual(fmt.Sprintf("%s-gpio%d", id, pin))
			return out, nil
		}
	})

	It("Should do nothing with a blank list of zones", func() {
		err := setupZones("", srv)
		Expect(err).To(BeNil())

		Expect(srv.Zones).To(HaveLen(0))
	})

	It("Should add zones with virtual outputs", func() {
		err := setupZones("foo:v,bar:v", srv)
		Expect(err).To(BeNil())

		Expect(srv.Zones).To(HaveLen(2))

		Expect(srv.Zones[0].ID).To(Equal("foo"))
		Expect(srv.Zones[1].ID).To(Equal("bar"))
		Expect(srv.Zones[0].Out.Id()).To(Equal("foo"))
		Expect(srv.Zones[1].Out.Id()).To(Equal("bar"))
	})

	It("Should add real outputs with correct pin", func() {
		err := setupZones("foo:10,bar:47", srv)
		Expect(err).To(BeNil())

		Expect(srv.Zones).To(HaveLen(2))

		Expect(srv.Zones[0].ID).To(Equal("foo"))
		Expect(srv.Zones[1].ID).To(Equal("bar"))
		Expect(srv.Zones[0].Out.Id()).To(Equal("foo-gpio10"))
		Expect(srv.Zones[1].Out.Id()).To(Equal("bar-gpio47"))
	})
})

var _ = Describe("Reading schedule from cmdline", func() {
	var (
		mockCtrl *gomock.Controller
		theTimer *mock_timer.MockTimer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		theTimer = mock_timer.NewMockTimer(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("Should add all given entries to the given timer", func() {

		schedule := "6:30,On;7:30,Off;19:30,On;21:00,Off"
		theTimer.EXPECT().AddEntry(6, 30, timer.TurnOn)
		theTimer.EXPECT().AddEntry(7, 30, timer.TurnOff)
		theTimer.EXPECT().AddEntry(19, 30, timer.TurnOn)
		theTimer.EXPECT().AddEntry(21, 0, timer.TurnOff)

		err := processCmdlineSchedule(schedule, theTimer)
		Expect(err).To(BeNil())
	})

	It("Should do nothing with a blank schedule", func() {
		err := processCmdlineSchedule("", theTimer)
		Expect(err).To(BeNil())
	})

	It("Should ignore a trailing ';'", func() {
		schedule := "6:30,On;7:30,Off;"
		theTimer.EXPECT().AddEntry(6, 30, timer.TurnOn)
		theTimer.EXPECT().AddEntry(7, 30, timer.TurnOff)

		err := processCmdlineSchedule(schedule, theTimer)
		Expect(err).To(BeNil())
	})

	Context("Error handling", func() {
		BeforeEach(func() {
			theTimer.EXPECT().AddEntry(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		})

		It("Should return an error with any invalid times", func() {
			err := processCmdlineSchedule("6:67,On;7:30,Off", theTimer)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("Invalid schedule entry 6:67,On"))

			err = processCmdlineSchedule("6:30,On;25:43,Off", theTimer)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("Invalid schedule entry 25:43,Off"))
		})

		It("Should return an error with any malformed parts", func() {
			err := processCmdlineSchedule("6:67,unsure;7:30,Off", theTimer)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("Invalid schedule entry 6:67,unsure"))

			err = processCmdlineSchedule("6:30,On;25-43_Off", theTimer)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("Invalid schedule entry 25-43_Off"))

			err = processCmdlineSchedule("6:30:45,On;25-43_Off", theTimer)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("Invalid schedule entry 6:30:45,On"))
		})
	})
})
