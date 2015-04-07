package zone

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alext/heating-controller/scheduler"
)

var DataDir string

func (o OverrideMode) MarshalText() ([]byte, error) {
	switch o {
	case ModeOverrideOn:
		return []byte("override_on"), nil
	case ModeOverrideOff:
		return []byte("override_off"), nil
	default:
		return []byte("normal"), nil
	}
}

func (o *OverrideMode) UnmarshalText(data []byte) error {
	switch string(data) {
	case "override_on":
		*o = ModeOverrideOn
	case "override_off":
		*o = ModeOverrideOff
	case "normal":
		*o = ModeNormal
	default:
		return fmt.Errorf("Unrecognised OverrideMode value '%s'", data)
	}
	return nil
}

type zoneData struct {
	OverrideMode OverrideMode      `json:"override_mode"`
	Events       []scheduler.Event `json:"events"`
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

	z.overrideMode = data.OverrideMode
	for _, e := range data.Events {
		err = z.Scheduler.AddEvent(e)
		if err != nil {
			log.Printf("[Zone:%s] Error restoring event '%v': %s", z.ID, e, err.Error())
		}
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

	data := zoneData{
		OverrideMode: z.overrideMode,
		Events:       z.Scheduler.ReadEvents(),
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("[Zone:%s] Error saving zone state: %s", z.ID, err.Error())
		return err
	}
	_, err = file.Write(b)
	if err != nil {
		log.Printf("[Zone:%s] Error saving zone state: %s", z.ID, err.Error())
		return err
	}

	return nil
}
