package integration_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"

	"github.com/alext/heating-controller/zone"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	log.SetOutput(ioutil.Discard)
	RunSpecs(t, "Integration tests")
}

var agoutiDriver *agouti.WebDriver

var _ = BeforeSuite(func() {
	var err error
	zone.DataDir, err = ioutil.TempDir("", "integration_test")
	Expect(err).NotTo(HaveOccurred())
	agoutiDriver = agouti.PhantomJS()
	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = AfterSuite(func() {
	os.RemoveAll(zone.DataDir)
	agoutiDriver.Stop()
})
