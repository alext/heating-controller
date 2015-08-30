package gpio

type pin struct {
	number int
	mode   Mode
	value  bool
}

func OpenPin(n int, mode Mode) (Pin, error) {
	p := &pin{
		number: n,
		mode:   mode,
	}
	return p, nil
}

func (p *pin) Close() error {
	return nil
}

func (p *pin) Mode() (Mode, error) {
	return p.mode, nil
}

func (p *pin) SetMode(mode Mode) error {
	p.mode = mode
	return nil
}

func (p *pin) Set() error {
	p.value = true
	return nil
}

func (p *pin) Clear() error {
	p.value = false
	return nil
}

func (p *pin) Get() (bool, error) {
	return p.value, nil
}

func (p *pin) BeginWatch(edge Edge, callback IRQEvent) error {
	panic("Watch is not yet implemented!")
}

func (p *pin) EndWatch() error {
	panic("Watch is not yet implemented!")
}

func (p *pin) Wait(condition bool) {
	panic("Wait is not yet implemented!")
}
