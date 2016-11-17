package thermostat

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type Thermostat interface {
	Current() Temperature
	Target() Temperature
	Set(Temperature)
	Close()
}

type demandFunc func(bool)

type thermostat struct {
	id      string
	url     string
	demand  demandFunc
	closeCh chan struct{}

	lock    sync.RWMutex
	target  Temperature
	current Temperature
	active  bool
}

func New(id string, url string, target Temperature, df demandFunc) Thermostat {

	t := &thermostat{
		id:      id,
		url:     url,
		target:  target,
		demand:  df,
		closeCh: make(chan struct{}),
	}

	// Set active so that a new thermostat defaults to active when within the
	// threshold.
	t.active = true
	t.readTemp()

	go t.readLoop()
	return t
}

func (t *thermostat) Current() Temperature {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.current
}

func (t *thermostat) Target() Temperature {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.target
}

func (t *thermostat) Set(tmp Temperature) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.target = tmp
	t.trigger()
}

func (t *thermostat) Close() {
	if t.closeCh != nil {
		close(t.closeCh)
	}
}

func (t *thermostat) readLoop() {
	tkr := time.NewTicker(time.Minute)
	defer tkr.Stop()
	for {
		select {
		case <-tkr.C:
			t.readTemp()
		case <-t.closeCh:
			return
		}
	}
}

func (t *thermostat) readTemp() {
	resp, err := http.Get(t.url)
	if err != nil {
		log.Printf("[Thermostat:%s] Error querying '%s': %s", t.id, t.url, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Thermostat:%s] Got %d querying '%s'", t.id, resp.StatusCode, t.url)
		return
	}

	var d struct {
		Temp *Temperature `json:"temperature"`
	}
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		log.Printf("[Thermostat:%s] Error decoding JSON from '%s': %s", t.id, t.url, err.Error())
		return
	}

	if d.Temp == nil {
		log.Printf("[Thermostat:%s] Missing temperature field in data from '%s'", t.id, t.url)
		return
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	t.current = *d.Temp
	t.trigger()
}

const threshold = 500

// Must be called with the lock held for writing.
func (t *thermostat) trigger() {
	if t.current < (t.target - threshold) {
		t.active = true
	} else if t.current > t.target { // no threshold here due to hysteresis in system.
		t.active = false
	}
	if t.demand != nil {
		go t.demand(t.active)
	}
}