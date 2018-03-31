package webserver_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/controller/controllerfakes"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/output/outputfakes"
	"github.com/alext/heating-controller/thermostat/mock_thermostat"
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

	Describe("changing an output state", func() {
		var (
			output1 output.Output
			zone1   *controller.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = controller.NewZone("one", output1)
			ctrl.AddZone(zone1)
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
				fakeOutput := new(outputfakes.FakeOutput)
				ctrl.AddZone(controller.NewZone("mock", fakeOutput))

				err := errors.New("Computer says no!")
				fakeOutput.ActivateReturns(err)

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
				fakeOutput := new(outputfakes.FakeOutput)
				ctrl.AddZone(controller.NewZone("mock", fakeOutput))

				err := errors.New("Computer says no!")
				fakeOutput.DeactivateReturns(err)

				w := doFakePutRequest(server, "/zones/mock/deactivate")

				Expect(w.Code).To(Equal(500))
				Expect(w.Body.String()).To(Equal("Error deactivating output 'mock': Computer says no!\n"))
			})
		})
	})

	Describe("boosting", func() {
		var (
			output1 output.Output
			zone1   *controller.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = controller.NewZone("one", output1)
			ctrl.AddZone(zone1)
		})

		Describe("setting the boost", func() {
			It("should boost the zone's scheduler", func() {
				fakeEventHandler := new(controllerfakes.FakeEventHandler)
				zone1.EventHandler = fakeEventHandler

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
				fakeEventHandler := new(controllerfakes.FakeEventHandler)
				zone1.EventHandler = fakeEventHandler

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
			BeforeEach(func() {
				zone1.Thermostat = mock_thermostat.New(19000)
			})

			It("increments the target and redirects back", func() {
				w := doRequest(server, "POST", "/zones/one/thermostat/increment")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))

				Expect(zone1.Thermostat.Target()).To(BeNumerically("==", 19500))
			})

			It("decrements the target and redirects back", func() {
				w := doRequest(server, "POST", "/zones/one/thermostat/decrement")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))

				Expect(zone1.Thermostat.Target()).To(BeNumerically("==", 18500))
			})

			It("saves the zone state", func() {
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
