package span

import (
	"time"
	"tracer/pkg/config"
)

type ToModel struct {
	Operation  string
	Context    SpanContext
	Tags       []config.Tag
	StartTime  time.Time
	Duration   time.Duration
	ProcessID  string
	References []Reference
	Logs       []Log
}
