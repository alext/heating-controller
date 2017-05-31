package sensor

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSensor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sensor")
}
