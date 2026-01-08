package tracer

import "tracer/pkg/span"

type Process struct {
	ServiceName string
	Tags        []span.Tag
}
