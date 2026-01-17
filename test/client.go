package main

import (
	"google.golang.org/grpc/metadata"
	"time"
	"tracer/pkg/config"
	"tracer/pkg/tracer"
)

func main() {
	t, err := tracer.NewTracer(&config.Configuration{
		ServiceName: "gateway",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			QueueSize: 5,
			Duration:  time.Second,
			AgentAddr: "127.0.0.1:8888",
		},
	}, config.Tag{
		Key:   "ip",
		Value: "localhost",
	})

	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second)

	span := t.StartSpan("HTTP GET /user/:id",
		tracer.WithTag("http.method", "GET"),
		tracer.WithTag("http.path", "/user/123"))

	time.Sleep(time.Second)

	span.SetTag("http.status_code", 200)
	span.SetBaggageItem("uid", "123456")
	span.LogFields(
		config.Tag{
			Key:   "event",
			Value: "cache_miss",
		})

	//fmt.Println(span)

	span.Finish()

	carrier := tracer.GRPCCarrier{
		MD: make(metadata.MD),
	}

	if err := t.Inject(span.Context, &carrier); err != nil {
		panic(err)
	}

	//fmt.Println(carrier)

	spanCtx, err := t.Extract(&carrier)
	if err != nil {
		panic(err)
	}

	//fmt.Println(spanCtx)
	childSpan := t.StartSpan("order_service", tracer.ChildOf(spanCtx))
	//fmt.Println(childSpan)

	childSpan.Finish()

	select {}
}
