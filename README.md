# Tracer

Tracer is a high-performance, distributed tracing system inspired by Uber's Jaeger, built from the ground up using Golang. It is designed to monitor and troubleshoot transactions in complex distributed systems.

## 🚀 Features

- **High Performance**: Built with Go for efficiency and low overhead.
- **Scalable Architecture**: Separates collection, processing, and storage.
- **ClickHouse Storage**: Utilizes ClickHouse for efficient storage and querying of trace data.
- **Kafka Buffering**: Uses Kafka for reliable data buffering and stream processing.
- **gRPC Communication**: Efficient communication between components using gRPC.

## 🏗 Architecture

The system consists of several key components:

- **Tracer SDK**: A library integrated into applications to generate trace spans. It handles sampling and reporting.
- **Agent**: A daemon that runs locally (or as a sidecar). It receives spans from the Tracer SDK, aggregates them, and sends them to the Collector.
- **Collector**: Receives traces from Agents, validates them, and pushes them to a Kafka topic.
- **Ingestor**: Consumes traces from Kafka and stores them into the ClickHouse database.

## 📦 Components

- `pkg/tracer`: The Go SDK for instrumenting applications.
- `cmd/agent.go`: The Agent service entry point.
- `internal/collector`: Logic for the Collector service.
- `internal/ingestor`: Logic for the Ingestor service.

## 🛠 Quick Start

### Installation

```bash
go get github.com/qing007-coder/tracer
```

### Usage (SDK)

Here is a simple example of how to use the Tracer SDK in your application:

```go
package main

import (
    "time"
    "tracer/pkg/config"
    "tracer/pkg/tracer"
)

func main() {
    // Initialize configuration
    conf := &config.Configuration{
        ServiceName: "my-service",
        // Add other config options...
    }

    // Create a new Tracer
    t, err := tracer.NewTracer(conf)
    if err != nil {
        panic(err)
    }

    // Start a new Span
    span := t.StartSpan("my-operation")
    defer span.Finish()

    // ... perform your operation ...
    time.Sleep(100 * time.Millisecond)
    
    // Add tags or logs
    span.SetTag("http.status_code", 200)
}
```

## 🗄 Storage Schema

The project includes a `clickhouse.sql` file which defines the database schema required for storing traces in ClickHouse.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
