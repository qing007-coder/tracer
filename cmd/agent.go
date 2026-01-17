package cmd

import (
	"fmt"
	"log"
	"time"
	"tracer/internal/agent"
	"tracer/pkg/model"
)

func NewAgent() {
	bufferToAggregator := make(chan model.Package, 512)
	aggregatorToExporter := make(chan model.BatchPackage, 200)
	buffer, err := agent.NewBuffer(":8888", bufferToAggregator)
	if err != nil {
		log.Println(err)
		return
	}
	aggregator := agent.NewAggregator(time.Second*2, bufferToAggregator, aggregatorToExporter, 1000)
	exporter, err := agent.NewExporter(10, aggregatorToExporter)
	if err != nil {
		log.Println(err)
		return
	}

	aggregator.Start()
	exporter.Start()
	if err := buffer.Listen(); err != nil {
		fmt.Println(err)
		return
	}
}
