package ingestor

import "tracer/pkg/model"

// Configuration holds the settings for the Ingestor service.
type Configuration struct {
	ConsumerToMergerChan []chan *model.FlatSpan
	MergerToStorageChan  chan []*model.StorageSpan
	WorkerNum            int
	BatchSize            int
}
