package tracer

import (
	"tracer/pkg/config"
	"tracer/pkg/span"
)

type Option interface {
	Apply(*StartSpanOption)
}

type StartSpanOption struct {
	TracerID   string
	References []span.Reference
	Tags       []config.Tag
	Baggage    map[string]string
}
