package zone

import (
	"encoding/json"
	"os"
	"syscall"

	"github.com/alext/heating-controller/scheduler"
)

var DataDir string

type EventAdder interface {
	AddEvent(scheduler.Event) error
}
type EventReader interface {
	ReadEvents() []scheduler.Event
}

type zoneData struct {
	Events []scheduler.Event `json:"events"`
}

func LoadEvents(id string, ea EventAdder) error {
	filename := DataDir + "/" + id + ".json"
	file, err := os.Open(filename)
	if err != nil {
		if perr, ok := err.(*os.PathError); ok {
			if perr.Err == syscall.ENOENT {
				return nil
			}
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
		err = ea.AddEvent(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveEvents(id string, er EventReader) error {
	filename := DataDir + "/" + id + ".json"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data := zoneData{Events: er.ReadEvents()}
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
