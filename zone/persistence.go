package zone

import (
	"encoding/json"
	"os"

	"github.com/alext/heating-controller/scheduler"
)

var DataDir string

type zoneData struct {
	Events []scheduler.Event `json:"events"`
}

func (z *Zone) Restore() error {
	filename := DataDir + "/" + z.ID + ".json"
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var data zoneData

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return err
	}
	for _, e := range data.Events {
		err = z.Scheduler.AddEvent(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (z *Zone) Save() error {
	filename := DataDir + "/" + z.ID + ".json"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data := zoneData{Events: z.Scheduler.ReadEvents()}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(b)
	if err != nil {
		return err
	}

	return nil
}
