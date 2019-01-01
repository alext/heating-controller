package webserver_test

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/controller/controllerfakes"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/thermostat/thermostatfakes"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("zones controller", func() {
	var (
		ctrl   *controller.Controller
		server *webserver.WebServer
	)

	BeforeEach(func() {
		ctrl = controller.New()
		server = webserver.New(ctrl, 8080, "")
	})

	Describe("JSON index", func() {
		BeforeEach(func() {
			out1 := output.Virtual("one")
			zone1 := controller.NewZone("one", out1)
			zone1.EventHandler = new(controllerfakes.FakeEventHandler)
			out2 := output.Virtual("two")
			zone2 := controller.NewZone("two", out2)
			zone2.EventHandler = new(controllerfakes.FakeEventHandler)
			ctrl.AddZone(zone1)
			ctrl.AddZone(zone2)
			out1.Activate()
		})

		It("returns details of the state of all zones", func() {
			resp := doGetRequest(server, "/zones")
			Expect(resp.Code).To(Equal(200))
			Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))

			data := decodeJsonResponse(resp)
			Expect(data).To(HaveKey("one"))
			Expect(data).To(HaveKey("two"))
			data1 := data["one"].(map[string]interface{})
			Expect(data1["active"]).To(BeTrue())
			data2 := data["two"].(map[string]interface{})
			Expect(data2["active"]).To(BeFalse())
		})
	})

	Describe("boosting", func() {
		var (
			output1          output.Output
			zone1            *controller.Zone
			fakeEventHandler *controllerfakes.FakeEventHandler
		)

		BeforeEach(func() {
			fakeEventHandler = new(controllerfakes.FakeEventHandler)
			output1 = output.Virtual("one")
			zone1 = controller.NewZone("one", output1)
			zone1.EventHandler = fakeEventHandler
			ctrl.AddZone(zone1)
		})

		Describe("setting the boost", func() {
			It("should boost the zone's scheduler", func() {
				doFakeRequestWithValues(server, "PUT", "/zones/one/boost", url.Values{"duration": {"42m"}})

				Expect(fakeEventHandler.BoostCallCount()).To(Equal(1))
				Expect(fakeEventHandler.BoostArgsForCall(0)).To(Equal(42 * time.Minute))
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
				doFakeDeleteRequest(server, "/zones/one/boost")

				Expect(fakeEventHandler.CancelBoostCallCount()).To(Equal(1))
			})

			It("should redirect to the index", func() {
				w := doFakeDeleteRequest(server, "/zones/one/boost")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))
			})
		})
	})

	Describe("incrementing/decrementing the thermostat", func() {
		var (
			tempDataDir string
			zone1       *controller.Zone
		)

		BeforeEach(func() {
			tempDataDir, _ = ioutil.TempDir("", "schedule_controller_test")
			controller.DataDir = tempDataDir
			zone1 = controller.NewZone("one", output.Virtual("one"))
			ctrl.AddZone(zone1)
		})

		AfterEach(func() {
			os.RemoveAll(tempDataDir)
		})

		Context("for a zone with a thermostat configured", func() {
			var (
				ts *thermostatfakes.FakeThermostat
			)

			BeforeEach(func() {
				ts = new(thermostatfakes.FakeThermostat)
				ts.TargetReturns(19000)
				zone1.Thermostat = ts
			})

			It("increments the target and redirects back", func() {
				w := doRequest(server, "POST", "/zones/one/thermostat/increment")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))

				Expect(ts.SetCallCount()).To(Equal(1))
				Expect(ts.SetArgsForCall(0)).To(BeNumerically("==", 19500))
			})

			It("decrements the target and redirects back", func() {
				w := doRequest(server, "POST", "/zones/one/thermostat/decrement")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))

				Expect(ts.SetCallCount()).To(Equal(1))
				Expect(ts.SetArgsForCall(0)).To(BeNumerically("==", 18500))
			})

			It("saves the zone state", func() {
				ts.TargetReturns(19500)

				doRequest(server, "POST", "/zones/one/thermostat/increment")

				file, err := os.Open(controller.DataDir + "/one.json")
				Expect(err).NotTo(HaveOccurred())
				var data map[string]interface{}
				err = json.NewDecoder(file).Decode(&data)
				Expect(err).NotTo(HaveOccurred())
				Expect(data["thermostat_target"]).To(BeNumerically("==", 19500))
			})
		})

		Context("for a zone without a thermostat configured", func() {

			It("should 404 on increment", func() {
				w := doRequest(server, "POST", "/zones/one/thermostat/increment")
				Expect(w.Code).To(Equal(404))
			})

			It("should 404 on decrement", func() {
				w := doRequest(server, "POST", "/zones/one/thermostat/decrement")
				Expect(w.Code).To(Equal(404))
			})
		})
	})
})
