package cmd

import (
	"fmt"
	"log"
	"time"
	"tracer/internal/agent"
	"tracer/pkg/model"
)

// NewAgent initializes and starts the agent components.
// It sets up the buffer, aggregator, and exporter to process trace data.
// The agent listens on port 8888 for incoming spans.
func NewAgent() {
	bufferToAggregator := make(chan model.Package, 512)
	aggregatorToExporter := make(chan model.BatchPackage, 200)

	// Create a new Buffer to receive spans
	buffer, err := agent.NewBuffer(":8888", bufferToAggregator)
	if err != nil {
		log.Println(err)
		return
	}

	// Create a new Aggregator to batch spans
	aggregator := agent.NewAggregator(time.Second*2, bufferToAggregator, aggregatorToExporter, 1000)

	// Create a new Exporter to send batches to the collector
	exporter, err := agent.NewExporter(10, aggregatorToExporter)
	if err != nil {
		log.Println(err)
		return
	}

	aggregator.Start()
	exporter.Start()

	// Start listening for incoming data
	if err := buffer.Listen(); err != nil {
		fmt.Println(err)
		return
	}
}
