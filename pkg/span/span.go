package span

import "time"

type Span struct {
	Operation  string
	Context    SpanContext
	Tags       []Tag
	StartTime  time.Time
	Duration   time.Duration
	ProcessID  string
	References []Reference
	Baggage    map[string]string
}
