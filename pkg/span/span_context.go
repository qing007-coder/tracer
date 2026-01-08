package span

type SpanContext struct {
	TraceID  string
	SpanID   string
	ParentID string
}
