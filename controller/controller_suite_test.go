package controller

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)

	RunSpecs(t, "Controller")
}

func todayAt(hour, minute, second int) time.Time {
	now := time.Now().Local()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())
}
