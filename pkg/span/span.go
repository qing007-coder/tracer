package span

import (
	"time"
	"tracer/pkg/config"
)

// Span represents a unit of work in a trace.
// It contains metadata such as operation name, start time, duration, tags, and logs.
type Span struct {
	Operation  string
	Context    SpanContext
	Tags       []config.Tag
	StartTime  time.Time
	Duration   time.Duration
	ProcessID  string
	References []Reference
	OnFinish   func(toModel *ToModel)
	Logs       []Log
}

// Finish marks the end of the span execution.
// It calculates the duration and triggers the OnFinish callback if the span is sampled.
func (s *Span) Finish() {
	s.Duration = time.Since(s.StartTime)
	if !s.Context.Sampled {
		return
	}

	s.OnFinish(s.ToModel())
}

// SetTag adds or updates a tag on the span.
// If the tag with the given key already exists, its value is updated.
func (s *Span) SetTag(key string, value interface{}) {
	for _, tag := range s.Tags {
		if tag.Key == key {
			tag.Value = value
			return
		}
	}

	s.Tags = append(s.Tags, config.Tag{
		Key:   key,
		Value: value,
	})
}

// SetBaggageItem sets a key:value pair on the span context that propagates to child spans.
func (s *Span) SetBaggageItem(key, value string) {
	s.Context.Baggage[key] = value
}

// GetBaggageItem retrieves the value of a baggage item from the span context.
func (s *Span) GetBaggageItem(key string) string {
	return s.Context.Baggage[key]
}

func (s *Span) LogFields(fields ...config.Tag) {
	var log Log
	for _, field := range fields {
		log.Fields = append(log.Fields, field)
	}

	log.Timestamp = time.Now()
	s.Logs = append(s.Logs, log)
}

func (s *Span) ToModel() *ToModel {
	return &ToModel{
		Operation:  s.Operation,
		Context:    s.Context,
		Tags:       s.Tags,
		StartTime:  s.StartTime,
		Duration:   s.Duration,
		ProcessID:  s.ProcessID,
		References: s.References,
		Logs:       s.Logs,
	}
}
