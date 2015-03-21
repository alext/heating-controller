package zone

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/scheduler"
)

var DataDir string

type zoneData struct {
	Events []scheduler.Event `json:"events"`
}

func (z *Zone) Restore() error {
	filename := filepath.Join(DataDir, z.ID+".json")
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("[Zone:%s] No saved state found for zone.", z.ID)
			return nil
		}
		return err
	}
	defer file.Close()

	var data zoneData

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		logger.Warnf("[Zone:%s] Error parsing saved zone state: %s", z.ID, err.Error())
		return err
	}
	for _, e := range data.Events {
		err = z.Scheduler.AddEvent(e)
		if err != nil {
			logger.Warnf("[Zone:%s] Error restoring event '%v': %s", z.ID, e, err.Error())
		}
	}
	return nil
}

func (z *Zone) Save() error {
	filename := filepath.Join(DataDir, z.ID+".json")
	file, err := os.Create(filename)
	if err != nil {
		logger.Warnf("[Zone:%s] Error saving zone state: %s", z.ID, err.Error())
		return err
	}
	defer file.Close()

	data := zoneData{Events: z.Scheduler.ReadEvents()}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logger.Warnf("[Zone:%s] Error saving zone state: %s", z.ID, err.Error())
		return err
	}

	_, err = file.Write(b)
	if err != nil {
		logger.Warnf("[Zone:%s] Error saving zone state: %s", z.ID, err.Error())
		return err
	}

	return nil
}
