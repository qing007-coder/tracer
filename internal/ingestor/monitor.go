package ingestor

import (
	"hash/fnv"
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
	signalChan []chan string
}

func NewMonitor(slotNum int, interval time.Duration, signalChan []chan string) *Monitor {
	m := new(Monitor)
	m.init(slotNum, interval, signalChan)
	return m
}

func (m *Monitor) init(slotNum int, interval time.Duration, signalChan []chan string) {
	m.currentPos = 0
	m.interval = interval
	m.slotNum = slotNum
	m.ticker = time.NewTicker(m.interval)
	m.slots = make([]map[string]struct{}, slotNum)
	m.signalChan = signalChan
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

	if len(data) == 0 {
		return
	}

	for traceID, _ := range data {
		id := m.route(traceID)
		m.signalChan[id] <- traceID
	}
}

func (m *Monitor) Add(traceID string, duration int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := (m.currentPos + duration) % m.slotNum
	m.slots[index][traceID] = struct{}{}
}

func (m *Monitor) route(traceID string) int {
	h := fnv.New32a()
	_, err := h.Write([]byte(traceID))
	if err != nil {
		return 0
	}
	return int(h.Sum32()) % len(m.signalChan)
}
