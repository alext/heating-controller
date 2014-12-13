package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/core"

	"github.com/alext/heating-controller/logger"
)

func TestWebServer(t *testing.T) {
	RegisterFailHandler(Fail)
	logger.SetDestination("/dev/null")
	RunSpecs(t, "Web Server Suite")
}

var agoutiDriver WebDriver

var _ = BeforeSuite(func() {
	var err error
	agoutiDriver, err = PhantomJS()
	Expect(err).NotTo(HaveOccurred())
	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = AfterSuite(func() {
	agoutiDriver.Stop()
})
