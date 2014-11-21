package output

import (
	"sync"
)

type virtual struct {
	id    string
	state bool
	mu    sync.Mutex
}

func Virtual(id string) Output {
	return &virtual{
		id: id,
	}
}

func (out *virtual) Id() string {
	return out.id
}

func (out *virtual) Active() (bool, error) {
	out.mu.Lock()
	defer out.mu.Unlock()
	return out.state, nil
}

func (out *virtual) Activate() error {
	out.mu.Lock()
	defer out.mu.Unlock()
	out.state = true
	return nil
}

func (out *virtual) Deactivate() error {
	out.mu.Lock()
	defer out.mu.Unlock()
	out.state = false
	return nil
}

func (out *virtual) Close() error {
	return nil
}
