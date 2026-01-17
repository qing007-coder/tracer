package ingestor

import (
	"sync"
	"time"
)

type Monitor struct {
	mu         sync.Mutex
	ticker     *time.Ticker
	slots      []map[string]struct{}
	currentPos int
	slotNum    int
	interval   time.Duration
	Handler    func(map[string]struct{})
}

func NewMonitor(slotNum int, interval time.Duration, handler func(map[string]struct{})) *Monitor {
	m := new(Monitor)
	m.init(slotNum, interval, handler)
	return m
}

func (m *Monitor) init(slotNum int, interval time.Duration, handler func(map[string]struct{})) {
	m.currentPos = 0
	m.interval = interval
	m.ticker = time.NewTicker(m.interval)
	m.slots = make([]map[string]struct{}, slotNum)
	m.Handler = handler
	for i := 0; i < slotNum; i++ {
		m.slots[i] = make(map[string]struct{})
	}
}

func (m *Monitor) Start() {
	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.Tick()
			}
		}

	}()
}

func (m *Monitor) Tick() {
	m.mu.Lock()

	data := m.slots[m.currentPos]
	m.slots[m.currentPos] = make(map[string]struct{})
	m.currentPos = (m.currentPos + 1) % m.slotNum
	m.mu.Unlock()

	if len(data) > 0 {
		go m.Handler(data)
	}
}

func (m *Monitor) Add(traceID string, position int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := (m.currentPos + position) % m.slotNum
	m.slots[index][traceID] = struct{}{}
}
