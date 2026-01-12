package span

type SpanContext struct {
	TraceID  string
	SpanID   string
	ParentID string
	Baggage  map[string]string
	Sampled  bool
}

func NewSpanContext() SpanContext {
	return SpanContext{
		Baggage: make(map[string]string),
	}
}

func (sc SpanContext) ForeachBaggageItem(handler func(k, v string)) {
	for k, v := range sc.Baggage {
		handler(k, v)
	}
}
