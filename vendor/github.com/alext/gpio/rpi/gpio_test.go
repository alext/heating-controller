package rpi

import (
	"testing"

	"github.com/alext/gpio"
)

func TestImplementsPin(t *testing.T) {
	var pin interface{} = new(pin)

	_, ok := pin.(gpio.Pin)
	if ! ok {
		t.Error("Expected pin to implement gpio.Pin")
	}
}
