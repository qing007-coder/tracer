package main

import (
	"fmt"
	"tracer/internal/ingestor"
	"tracer/pkg/model"
)

func main() {
	consumerToMergerChan := make([]chan *model.FlatSpan, 0)
	workNum := 5
	for i := 0; i < workNum; i++ {
		consumerToMergerChan = append(consumerToMergerChan, make(chan *model.FlatSpan, 1000))
	}
	conf := ingestor.Configuration{
		ConsumerToMergerChan: consumerToMergerChan,
		MergerToStorageChan:  make(chan []*model.StorageSpan, 100),
		WorkerNum:            workNum,
		BatchSize:            10000,
	}

	consumer, err := ingestor.NewConsumer(conf)
	if err != nil {
		fmt.Println("errL:", err)
		return
	}

	go consumer.Run()

	merger := ingestor.NewMerger(conf)
	merger.Start()

	storage := ingestor.NewStorage(conf)
	storage.Run()
}
