package webserver_test

import (
	"errors"
	"net/url"
	"time"

	"code.google.com/p/gomock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/output/mock_output"
	"github.com/alext/heating-controller/scheduler/mock_scheduler"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

var _ = Describe("zones controller", func() {
	var (
		mockCtrl *gomock.Controller
		server   *webserver.WebServer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		server = webserver.New(8080)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("changing an output state", func() {
		var (
			output1 output.Output
			zone1   *zone.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = zone.New("one", output1)
			server.AddZone(zone1)
		})

		Describe("activating the zone's output", func() {

			It("should activate the output", func() {
				doFakePutRequest(server, "/zones/one/activate")

				Expect(output1.Active()).To(Equal(true))
			})

			It("should redirect to the index", func() {
				w := doFakePutRequest(server, "/zones/one/activate")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))
			})

			It("should show an error if activating fails", func() {
				mockOutput := mock_output.NewMockOutput(mockCtrl)
				server.AddZone(zone.New("mock", mockOutput))

				err := errors.New("Computer says no!")
				mockOutput.EXPECT().Activate().Return(err)

				w := doFakePutRequest(server, "/zones/mock/activate")

				Expect(w.Code).To(Equal(500))
				Expect(w.Body.String()).To(Equal("Error activating output 'mock': Computer says no!\n"))
			})
		})

		Describe("deactivating the zone's output", func() {
			BeforeEach(func() {
				output1.Activate()
			})

			It("should deactivate the output", func() {
				doFakePutRequest(server, "/zones/one/deactivate")

				Expect(output1.Active()).To(Equal(false))
			})

			It("should redirect to the index", func() {
				w := doFakePutRequest(server, "/zones/one/deactivate")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))
			})

			It("should show an error if activating fails", func() {
				mockOutput := mock_output.NewMockOutput(mockCtrl)
				server.AddZone(zone.New("mock", mockOutput))

				err := errors.New("Computer says no!")
				mockOutput.EXPECT().Deactivate().Return(err)

				w := doFakePutRequest(server, "/zones/mock/deactivate")

				Expect(w.Code).To(Equal(500))
				Expect(w.Body.String()).To(Equal("Error deactivating output 'mock': Computer says no!\n"))
			})
		})
	})

	Describe("boosting", func() {
		var (
			output1 output.Output
			zone1   *zone.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = zone.New("one", output1)
			server.AddZone(zone1)
		})

		Describe("setting the boost", func() {
			It("should boost the zone's scheduler", func() {
				mockScheduler := mock_scheduler.NewMockScheduler(mockCtrl)
				zone1.Scheduler = mockScheduler

				mockScheduler.EXPECT().Boost(42 * time.Minute)

				doFakeRequestWithValues(server, "PUT", "/zones/one/boost", url.Values{"duration": {"42m"}})
			})

			It("should redirect to the index", func() {
				w := doFakeRequestWithValues(server, "PUT", "/zones/one/boost", url.Values{"duration": {"42m"}})

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))
			})

			It("should return an error with an invalid duration", func() {
				w := doFakeRequestWithValues(server, "PUT", "/zones/one/boost", url.Values{"duration": {"wibble"}})
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(Equal("Invalid boost duration 'wibble'\n"))

				w = doFakeRequestWithValues(server, "PUT", "/zones/one/boost", url.Values{"duration": {""}})
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(Equal("Invalid boost duration ''\n"))
			})
		})

		Describe("cancelling the boost", func() {
			It("should boost the zone's scheduler", func() {
				mockScheduler := mock_scheduler.NewMockScheduler(mockCtrl)
				zone1.Scheduler = mockScheduler

				mockScheduler.EXPECT().CancelBoost()

				doFakeDeleteRequest(server, "/zones/one/boost")
			})

			It("should redirect to the index", func() {
				w := doFakeDeleteRequest(server, "/zones/one/boost")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))
			})
		})
	})

})
