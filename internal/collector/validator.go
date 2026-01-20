package collector

import (
	"fmt"
	"time"
	pb "tracer/internal/proto"
)

// Validator validates the structure and content of trace packages.
type Validator struct {
}

// NewValidator creates a new Validator.
func NewValidator() (*Validator, error) {
	return &Validator{}, nil
}

// Validate checks the batch package for missing fields, invalid time, or excessive size.
func (v *Validator) Validate(pkg *pb.BatchPackage) error {
	for _, pkg := range pkg.GetPackages() {
		// 1. Validate Process
		if pkg.GetProcess().GetServiceName() == "" {
			return fmt.Errorf("missing service name")
		}

		// 2. Validate Spans
		for _, s := range pkg.GetSpans() {
			ctx := s.GetContext()
			if ctx.GetTraceId() == "" || ctx.GetSpanId() == "" {
				return fmt.Errorf("missing trace/span id")
			}

			// 3. Time validity (check for future timestamps or too old)
			startTime := s.GetStartTime().AsTime()
			if time.Since(startTime) > 24*time.Hour {
				return fmt.Errorf("span is too old")
			}

			// 4. Size limits (prevent large tags from clogging Kafka)
			for k, v := range s.GetTags() {
				if len(k) > 128 || len(v) > 2048 {
					return fmt.Errorf("tag too large")
				}
			}
		}
	}
	return nil
}
