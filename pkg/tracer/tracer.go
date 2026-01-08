package tracer

import "tracer/pkg/span"

type Tracer struct {
	TraceID   string
	Spans     []*span.Span
	Processes map[string]Process
}
