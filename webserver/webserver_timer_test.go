package webserver_test

import (

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.google.com/p/gomock/gomock"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/timer/mock_timer"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("Timer API", func() {
	var (
		mockCtrl *gomock.Controller
		ctrl	  controller.Controller
		server   *webserver.WebServer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctrl = controller.New()
		server = webserver.New(ctrl, 8080)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("timers index", func() {
		It("should return an empty list of timers as json", func() {
			w := doGetRequest(server, "/timers")

			Expect(w.Code).To(Equal(200))
			Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(w.Body.String()).To(Equal("{}"))
		})

		PContext("with some timers", func() {
			It("should return a list of timers with basic info", func() {
			})
		})
	})

	Describe("returning details of a timer", func() {
		var (
			timer1 *mock_timer.MockTimer
		)

		BeforeEach(func() {
			timer1 = mock_timer.NewMockTimer(mockCtrl)
			timer1.EXPECT().Id().AnyTimes().Return("one")
			timer1.EXPECT().OutputActive().AnyTimes().Return(true)
			ctrl.AddTimer(timer1)
		})

		It("should return details of the requested timer", func() {
			w := doGetRequest(server, "/timers/one")

			Expect(w.Code).To(Equal(200))
			Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))

			data := decodeJsonResponse(w)
			Expect(data["id"]).To(Equal("one"))
			Expect(data["output_active"]).To(Equal(true))
		})

		It("should 404 for a non-existent timer", func() {
			w := doGetRequest(server, "/timers/non-existent")

			Expect(w.Code).To(Equal(404))
		})
	})
})
