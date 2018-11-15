package controller

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/alext/heating-controller/units"
)

var DataDir string

type zoneData struct {
	Events           []Event            `json:"events"`
	ThermostatTarget *units.Temperature `json:"thermostat_target,omitempty"`
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
	err = handleLegacyFormat(&data, file)
	if err != nil {
		log.Printf("[Zone:%s] Error handling legacy format: %s", z.ID, err.Error())
		return err
	}
	for _, e := range data.Events {
		err = z.AddEvent(e)
		if err != nil {
			log.Printf("[Zone:%s] Error restoring event '%v': %s", z.ID, e, err.Error())
		}
	}
	if data.ThermostatTarget != nil && z.Thermostat != nil {
		z.Thermostat.Set(*data.ThermostatTarget)
	}
	return nil
}

// FIXME: remove all this once we've migrated everywhere.
type legacyEvent struct {
	Hour        int               `json:"hour"`
	Min         int               `json:"min"`
	Action      Action            `json:"action"`
	ThermAction *ThermostatAction `json:"therm_action,omitempty"`
}

type legacyZoneData struct {
	Events []legacyEvent `json:"events"`
}

func handleLegacyFormat(data *zoneData, file io.ReadSeeker) error {
	if len(data.Events) == 0 || data.Events[0].Time != 0 {
		// nothing to do
		return nil
	}

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	var legacyData legacyZoneData
	err = json.NewDecoder(file).Decode(&legacyData)
	if err != nil {
		return err
	}
	if len(legacyData.Events) == 0 {
		return nil
	}
	if legacyData.Events[0].Hour == 0 && legacyData.Events[0].Min == 0 {
		// looks like no logacy data present, so abort.
		return nil
	}
	data.Events = make([]Event, 0, len(legacyData.Events))
	for _, e := range legacyData.Events {
		data.Events = append(data.Events, Event{
			Time:        units.NewTimeOfDay(e.Hour, e.Min),
			Action:      e.Action,
			ThermAction: e.ThermAction,
		})
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

	data := zoneData{Events: z.ReadEvents()}
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
