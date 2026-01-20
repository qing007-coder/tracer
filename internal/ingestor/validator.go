package ingestor

import (
	"fmt"
	"tracer/pkg/model"
)

// Validator checks if the FlatSpan has all necessary fields before processing.
type Validator struct {
}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate ensures the span has TraceID, SpanID, ServiceName and a valid StartTime.
func (v *Validator) Validate(span *model.FlatSpan) error {
	if span.TraceID == "" || span.SpanID == "" {
		return fmt.Errorf("missing essential IDs")
	}
	if span.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if span.StartTime <= 0 {
		return fmt.Errorf("invalid start time")
	}
	return nil
}
