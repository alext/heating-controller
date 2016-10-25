package thermostat

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestThermostat(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)

	RunSpecs(t, "Thermostat")
}

var _ = Describe("A Thermostat", func() {

	Describe("setting the target temperature", func() {
		var t *thermostat
		BeforeEach(func() {
			t = &thermostat{
				current: 18000,
				target:  17000,
			}
		})

		It("Updates the target for the thermostat", func() {
			t.Set(16000)
			Expect(t.target).To(BeNumerically("==", 16000))
		})

		It("triggers the thermostat to update the demand state", func() {
			t.Set(19000)
			Expect(t.active).To(BeTrue())
		})
	})

	Describe("reading the temperature from the source", func() {
		var (
			server *httptest.Server
			t      *thermostat
		)
		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("happy path", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					data := map[string]interface{}{"temperature": 19000}
					Expect(json.NewEncoder(w).Encode(&data)).To(Succeed())
				}))
				t = &thermostat{
					url:     server.URL,
					current: 17000,
					target:  18000,
					active:  true,
				}
			})

			It("updates the current temperature", func() {
				t.readTemp()
				Expect(t.current).To(BeNumerically("==", 19000))
			})

			It("triggers the thermostat to update the demand state", func() {
				t.readTemp()
				Expect(t.active).To(BeFalse())
			})
		})

		Describe("error handling", func() {
			BeforeEach(func() {
				t = &thermostat{
					current: 18000,
				}
			})

			Context("when the network connection fails", func() {
				It("doesn't update the current temperature", func() {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
					t.url = server.URL
					server.Close()
					t.readTemp()
					Expect(t.current).To(BeNumerically("==", 18000))
				})
			})

			Context("when the HTTP returns non-200 response", func() {
				It("doesn't update the current temperature", func() {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
						data := map[string]interface{}{"temperature": 19000}
						Expect(json.NewEncoder(w).Encode(&data)).To(Succeed())
					}))
					t.url = server.URL
					t.readTemp()
					Expect(t.current).To(BeNumerically("==", 18000))
				})
			})

			Context("when the response isn't JSON", func() {
				It("doesn't update the current temperature", func() {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("I'm not JSON..."))
					}))
					t.url = server.URL
					t.readTemp()
					Expect(t.current).To(BeNumerically("==", 18000))
				})
			})

			Context("when the response does not include a temperature field", func() {
				It("doesn't update the current temperature", func() {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						data := map[string]interface{}{"something": "else"}
						Expect(json.NewEncoder(w).Encode(&data)).To(Succeed())
					}))
					t.url = server.URL
					t.readTemp()
					Expect(t.current).To(BeNumerically("==", 18000))
				})
			})
		})
	})

	type TriggeringCase struct {
		Current         Temperature
		Target          Temperature
		CurrentlyActive bool
		ExpectedActive  bool
	}

	DescribeTable("triggering changes in state",
		func(c TriggeringCase) {
			demandCalled := false
			demandNotify := make(chan struct{})
			t := &thermostat{
				current: c.Current,
				target:  c.Target,
				active:  c.CurrentlyActive,
				demand: func(param bool) {
					demandCalled = true
					Expect(param).To(Equal(c.ExpectedActive))
					close(demandNotify)
				},
			}

			t.trigger()
			Expect(t.active).To(Equal(c.ExpectedActive))
			<-demandNotify
			Expect(demandCalled).To(BeTrue(), "expected demandFunc to be called")
		},
		Entry("activates when current well below target", TriggeringCase{
			Current: 15000, Target: 18000, CurrentlyActive: false,
			ExpectedActive: true,
		}),
		Entry("remains active when current well below target", TriggeringCase{
			Current: 15000, Target: 18000, CurrentlyActive: true,
			ExpectedActive: true,
		}),
		Entry("deactivates when current well above target", TriggeringCase{
			Current: 20000, Target: 18000, CurrentlyActive: true,
			ExpectedActive: false,
		}),
		Entry("remains inactive when current well above target", TriggeringCase{
			Current: 20000, Target: 18000, CurrentlyActive: false,
			ExpectedActive: false,
		}),
		Entry("remains active when current within threhold below target", TriggeringCase{
			Current: 17950, Target: 18000, CurrentlyActive: true,
			ExpectedActive: true,
		}),
		Entry("remains inactive when current within threhold below target", TriggeringCase{
			Current: 17950, Target: 18000, CurrentlyActive: false,
			ExpectedActive: false,
		}),
		Entry("remains active when current within threhold above target", TriggeringCase{
			Current: 18050, Target: 18000, CurrentlyActive: true,
			ExpectedActive: true,
		}),
		Entry("remains inactive when current within threhold above target", TriggeringCase{
			Current: 18050, Target: 18000, CurrentlyActive: false,
			ExpectedActive: false,
		}),
	)
})
