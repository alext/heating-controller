package main

import (
	"encoding/json"

	"github.com/alext/afero"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parsing the config file", func() {
	BeforeEach(func() {
		fs = &afero.MemMapFs{}
	})

	Context("with a config file", func() {

		It("should set the port", func() {
			createConfigFile("/etc/config.json", configData{
				"port": 1234,
			})

			config, err := loadConfig("/etc/config.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Port).To(Equal(1234))
		})

		It("should set a default port if none is specified", func() {
			createConfigFile("/etc/config.json", configData{})

			config, err := loadConfig("/etc/config.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Port).To(Equal(defaultPort))
		})

		It("should setup the zone details", func() {
			createConfigFile("/etc/config.json", configData{
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

			config, err := loadConfig("/etc/config.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Zones).To(HaveLen(3))

			Expect(config.Zones["foo"].GPIOPin).To(Equal(42))
			Expect(config.Zones["foo"].Virtual).To(BeFalse())
			Expect(config.Zones["bar"].GPIOPin).To(Equal(12))
			Expect(config.Zones["bar"].Virtual).To(BeFalse())
			Expect(config.Zones["baz"].Virtual).To(BeTrue())
		})

		It("should have an empty list of zones if none given", func() {
			createConfigFile("/etc/config.json", configData{})

			config, err := loadConfig("/etc/config.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Zones).To(HaveLen(0))
		})
	})

	Context("when the config file doesn't exist", func() {
		It("should set a default port", func() {
			config, err := loadConfig("/etc/config.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Port).To(Equal(defaultPort))
		})

		It("should set an empty list of zones", func() {
			config, err := loadConfig("/etc/config.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Zones).To(HaveLen(0))
		})
	})
})

type configData map[string]interface{}

func createConfigFile(filename string, data configData) {
	file, err := fs.Create(filename)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	defer file.Close()
	err = json.NewEncoder(file).Encode(data)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}
