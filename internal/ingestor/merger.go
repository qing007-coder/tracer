package ingestor

import (
	"fmt"
	"time"
	"tracer/pkg/model"
	"tracer/pkg/utils"
)

// Merger receives spans from consumers and merges them by trace ID.
// It caches spans in memory until a signal (e.g., timeout) triggers output.
type Merger struct {
	spanCache  []map[string][]*model.StorageSpan
	spanChan   []chan *model.FlatSpan
	workerNum  int
	signalChan []chan string // 用于monitor和merger进行通信的channel 单向
	monitor    *Monitor
	outputChan chan []*model.StorageSpan
}

// NewMerger creates a new Merger.
func NewMerger(conf Configuration) *Merger {
	m := new(Merger)
	m.init(conf)

	return m
}

// init initializes the Merger with configuration and channels.
func (m *Merger) init(conf Configuration) {
	m.workerNum = conf.WorkerNum
	m.spanCache = make([]map[string][]*model.StorageSpan, 0)
	m.outputChan = conf.MergerToStorageChan
	m.spanChan = conf.ConsumerToMergerChan

	signalChan := make([]chan string, 0)
	for i := 0; i < m.workerNum; i++ {
		m.spanCache = append(m.spanCache, make(map[string][]*model.StorageSpan))
		signalChan = append(signalChan, make(chan string, 10))
	}

	m.signalChan = signalChan

	m.monitor = NewMonitor(60, time.Second, signalChan)
}

// Start launches the monitor and worker goroutines.
func (m *Merger) Start() {
	m.monitor.Start()

	for i := 0; i < m.workerNum; i++ {
		go m.WorkerRun(i)
	}
}

// WorkerRun processes spans for a specific worker ID.
// It listens for new spans to store and signals to output merged traces.
func (m *Merger) WorkerRun(id int) {
	spanChan := m.spanChan[id]
	signalChan := m.signalChan[id]
	cache := m.spanCache[id]

	for {
		select {
		case span, ok := <-spanChan:
			if !ok {
				fmt.Println("out!!!!!")
				return
			}
			m.Store(span, cache)
		case traceID := <-signalChan:
			m.output(traceID, cache)
		}
	}
}

// Store adds a span to the local cache and notifies the monitor.
func (m *Merger) Store(span *model.FlatSpan, cache map[string][]*model.StorageSpan) {

	storageSpan := utils.FlatSpanToClickHouseSpan(span)
	spans, ok := cache[storageSpan.TraceID]
	if !ok {
		spans = make([]*model.StorageSpan, 0)
		cache[storageSpan.TraceID] = spans
	}

	spans = append(spans, storageSpan)
	cache[storageSpan.TraceID] = spans

	m.monitor.Add(storageSpan.TraceID, 5)
}

// output sends the cached spans for a given trace ID to the output channel.
func (m *Merger) output(traceID string, cache map[string][]*model.StorageSpan) {
	snapshot, ok := cache[traceID]
	if !ok || len(snapshot) == 0 {
		return
	}

	delete(cache, traceID) // 彻底移除，防止内存缓慢增长
	m.outputChan <- snapshot
}
