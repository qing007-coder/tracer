package ingestor

import "tracer/pkg/model"

type Configuration struct {
	ConsumerToMergerChan []chan *model.FlatSpan
	MergerToStorageChan  []chan *model.StorageSpan
	WorkerNum            int
}
