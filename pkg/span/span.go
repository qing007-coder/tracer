package span

import "time"

type Span struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Operation string
	Tags      []Tag
	StartTime time.Time
	duration  time.Duration
}
