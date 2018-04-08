package config_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"testing"

	"github.com/alext/heating-controller/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	log.SetOutput(ioutil.Discard)
	RunSpecs(t, "Config")
}

var _ = Describe("Parsing the config data", func() {
	var (
		configReader io.Reader
	)

	Context("with a config file", func() {

		It("should set the port", func() {
			configReader = createConfigReader(configData{
				"port": 1234,
			})

			cfg, err := config.LoadConfig(configReader)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Port).To(Equal(1234))
		})

		It("should set a default port if none is specified", func() {
			configReader = createConfigReader(configData{})

			cfg, err := config.LoadConfig(configReader)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Port).To(Equal(config.DefaultPort))
		})

		It("should setup the sensor details", func() {
			configReader = createConfigReader(configData{
				"sensors": map[string]map[string]interface{}{
					"foo": {
						"type": "w1",
						"id":   "1234",
					},
					"bar": {
						"type": "push",
						"id":   "2345",
					},
				},
			})

			cfg, err := config.LoadConfig(configReader)

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Sensors).To(HaveLen(2))

			Expect(cfg.Sensors["foo"].Type).To(Equal("w1"))
			Expect(cfg.Sensors["foo"].ID).To(Equal("1234"))
			Expect(cfg.Sensors["bar"].Type).To(Equal("push"))
			Expect(cfg.Sensors["bar"].ID).To(Equal("2345"))
		})

		It("should setup the zone details", func() {
			configReader = createConfigReader(configData{
				"zones": map[string]map[string]interface{}{
					"foo": {
						"gpio_pin": 42,
					},
					"bar": {
						"gpio_pin": 12,
					},
					"baz": {
						"virtual": true,
					},
				},
			})

			cfg, err := config.LoadConfig(configReader)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Zones).To(HaveLen(3))

			Expect(cfg.Zones["foo"].GPIOPin).To(Equal(42))
			Expect(cfg.Zones["foo"].Virtual).To(BeFalse())
			Expect(cfg.Zones["bar"].GPIOPin).To(Equal(12))
			Expect(cfg.Zones["bar"].Virtual).To(BeFalse())
			Expect(cfg.Zones["baz"].Virtual).To(BeTrue())
		})

		It("should have an empty list of sensors and zones if none given", func() {
			configReader = createConfigReader(configData{})

			cfg, err := config.LoadConfig(configReader)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Sensors).To(HaveLen(0))
			Expect(cfg.Zones).To(HaveLen(0))
		})

		Describe("adding thermostat details", func() {

			It("should add thermostat details if present", func() {
				configReader = createConfigReader(configData{
					"zones": map[string]map[string]interface{}{
						"foo": {
							"gpio_pin": 42,
							"thermostat": map[string]interface{}{
								"sensor":         "foo",
								"default_target": 18000,
							},
						},
					},
				})

				cfg, err := config.LoadConfig(configReader)
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.Zones["foo"].Thermostat.Sensor).To(Equal("foo"))
				Expect(cfg.Zones["foo"].Thermostat.DefaultTarget).To(BeNumerically("==", 18000))
			})

			It("should set thermostat to nil if no details present", func() {
				configReader = createConfigReader(configData{
					"zones": map[string]map[string]interface{}{
						"foo": {
							"gpio_pin": 42,
						},
					},
				})

				cfg, err := config.LoadConfig(configReader)
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.Zones["foo"].Thermostat).To(BeNil())
			})
		})
	})
})

type configData map[string]interface{}

func createConfigReader(data configData) io.Reader {
	var out bytes.Buffer
	err := json.NewEncoder(&out).Encode(data)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return &out
}
