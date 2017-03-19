package controller

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/alext/heating-controller/scheduler"
	"github.com/alext/heating-controller/sensor"
)

var DataDir string

type zoneData struct {
	Events           []scheduler.Event   `json:"events"`
	ThermostatTarget *sensor.Temperature `json:"thermostat_target,omitempty"`
}

func (z *Zone) Restore() error {
	filename := filepath.Join(DataDir, z.ID+".json")
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[Zone:%s] No saved state found for zone.", z.ID)
			return nil
		}
		return err
	}
	defer file.Close()

	var data zoneData

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		log.Printf("[Zone:%s] Error parsing saved zone state: %s", z.ID, err.Error())
		return err
	}
	for _, e := range data.Events {
		err = z.Scheduler.AddEvent(e)
		if err != nil {
			log.Printf("[Zone:%s] Error restoring event '%v': %s", z.ID, e, err.Error())
		}
	}
	if data.ThermostatTarget != nil && z.Thermostat != nil {
		z.Thermostat.Set(*data.ThermostatTarget)
	}
	return nil
}

func (z *Zone) Save() error {
	filename := filepath.Join(DataDir, z.ID+".json")
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("[Zone:%s] Error saving zone state: %s", z.ID, err.Error())
		return err
	}
	defer file.Close()

	data := zoneData{Events: z.Scheduler.ReadEvents()}
	if z.Thermostat != nil {
		// temporary variable needed so we can take the address of it.
		target := z.Thermostat.Target()
		data.ThermostatTarget = &target
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(data)
	if err != nil {
		log.Printf("[Zone:%s] Error saving zone state: %s", z.ID, err.Error())
		return err
	}

	return nil
}
