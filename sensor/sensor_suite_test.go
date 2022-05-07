package sensor

import (
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSensor(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(GinkgoWriter)

	RunSpecs(t, "Sensor")
}
