package main

import (
	"fmt"
	"tracer/internal/ingestor"
)

func main() {
	consumer, err := ingestor.NewConsumer()
	if err != nil {
		fmt.Println("errL:", err)
		return
	}

	consumer.Run()
}
