package dummy_timer


import (
	"sort"
	"sync"

	"github.com/alext/heating-controller/timer"
)

type entry struct {
	t      *dummyTimer
	hour   int
	min    int
	action timer.Action
}

type entryList []*entry

func (el entryList) Len() int      { return len(el) }
func (el entryList) Swap(i, j int) { el[i], el[j] = el[j], el[i] }
func (el entryList) Less(i, j int) bool {
	a, b := el[i], el[j]
	return a.hour < b.hour || (a.hour == b.hour && a.min < b.min)
}

type dummyTimer struct {
	entries      entryList
	running      bool
	lock         sync.Mutex
}

func New() timer.Timer {
	return &dummyTimer{
		entries:  make(entryList, 0),
	}
}

func (t *dummyTimer) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.running = true
}

func (t *dummyTimer) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.running = false
}

func (t *dummyTimer) Running() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.running
}

func (t *dummyTimer) AddEntry(hour, min int, a timer.Action) {
	t.lock.Lock()
	defer t.lock.Unlock()
	e := &entry{t: t, hour: hour, min: min, action: a}
	t.entries = append(t.entries, e)
	sort.Sort(t.entries)
}
