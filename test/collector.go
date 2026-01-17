package main

import (
	"fmt"
	"time"
	"tracer/internal/collector"
	"tracer/pkg/model"
)

func main() {
	receiverToBatcher := make(chan *model.FlatSpan, 1000)
	batcherToDispatcher := make(chan []*model.FlatSpan, 10)
	receiver, err := collector.NewReceiver(receiverToBatcher)
	if err != nil {
		fmt.Println(err)
		return
	}

	batcher, err := collector.NewBatcher(1000, receiverToBatcher, batcherToDispatcher, time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	batcher.Start()

	dispatcher, err := collector.NewDispatcher(10, batcherToDispatcher)
	if err != nil {
		fmt.Println(err)
		return
	}

	dispatcher.Start()

	receiver.Run()
}
