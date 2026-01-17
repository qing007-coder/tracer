package ingestor

import (
	"fmt"
	"tracer/pkg/model"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

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
