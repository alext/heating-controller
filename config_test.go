package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parsing the config file", func() {
	var (
		configFileName string
	)

	AfterEach(func() {
		if configFileName != "" {
			os.Remove(configFileName)
		}
	})

	Context("with a config file", func() {

		It("should set the port", func() {
			configFileName = createConfigFile(configData{
				"port": 1234,
			})

			config, err := loadConfig(configFileName)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Port).To(Equal(1234))
		})

		It("should set a default port if none is specified", func() {
			configFileName = createConfigFile(configData{})

			config, err := loadConfig(configFileName)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Port).To(Equal(defaultPort))
		})

		It("should setup the zone details", func() {
			configFileName = createConfigFile(configData{
				"zones": map[string]map[string]interface{}{
					"foo": map[string]interface{}{
						"gpio_pin": 42,
					},
					"bar": map[string]interface{}{
						"gpio_pin": 12,
					},
					"baz": map[string]interface{}{
						"virtual": true,
					},
				},
			})

			config, err := loadConfig(configFileName)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Zones).To(HaveLen(3))

			Expect(config.Zones["foo"].GPIOPin).To(Equal(42))
			Expect(config.Zones["foo"].Virtual).To(BeFalse())
			Expect(config.Zones["bar"].GPIOPin).To(Equal(12))
			Expect(config.Zones["bar"].Virtual).To(BeFalse())
			Expect(config.Zones["baz"].Virtual).To(BeTrue())
		})

		It("should have an empty list of zones if none given", func() {
			configFileName = createConfigFile(configData{})

			config, err := loadConfig(configFileName)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Zones).To(HaveLen(0))
		})

		Describe("adding thermostat details", func() {

			It("should add thermostat details if present", func() {
				configFileName = createConfigFile(configData{
					"zones": map[string]map[string]interface{}{
						"foo": map[string]interface{}{
							"gpio_pin": 42,
							"thermostat": map[string]interface{}{
								"sensor_url":     "http://foo.example.com/sensors/foo",
								"default_target": 18000,
							},
						},
					},
				})

				config, err := loadConfig(configFileName)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Zones["foo"].Thermostat.SensorURL).To(Equal("http://foo.example.com/sensors/foo"))
				Expect(config.Zones["foo"].Thermostat.DefaultTarget).To(BeNumerically("==", 18000))
			})

			It("should set thermostat to nil if no details present", func() {
				configFileName = createConfigFile(configData{
					"zones": map[string]map[string]interface{}{
						"foo": map[string]interface{}{
							"gpio_pin": 42,
						},
					},
				})

				config, err := loadConfig(configFileName)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Zones["foo"].Thermostat).To(BeNil())
			})
		})
	})

	Context("when the config file doesn't exist", func() {
		It("should set a default port", func() {
			config, err := loadConfig("/non-existent.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Port).To(Equal(defaultPort))
		})

		It("should set an empty list of zones", func() {
			config, err := loadConfig("/non-existent.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Zones).To(HaveLen(0))
		})
	})
})

type configData map[string]interface{}

func createConfigFile(data configData) string {
	file, err := ioutil.TempFile("", "config")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	defer file.Close()
	err = json.NewEncoder(file).Encode(data)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return file.Name()
}
