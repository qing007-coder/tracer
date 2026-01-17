package ingestor

import (
	"sync"
	"tracer/pkg/model"
)

type Merger struct {
	mu        sync.Mutex
	spanCache []map[string][]*model.StorageSpan
	spanChan  []chan *model.FlatSpan
	workerNum int
}
