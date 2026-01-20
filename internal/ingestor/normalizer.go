package ingestor

import (
	"strings"
	"time"
	"tracer/pkg/model"
)

// Normalizer standardizes span data (e.g., service names, timestamps).
type Normalizer struct {
}

// NewNormalizer creates a new Normalizer.
func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// Normalize modifies the span in place to ensure consistency.
// It lowercases service names, adds default tags, and standardizes timestamps.
func (n *Normalizer) Normalize(span *model.FlatSpan) {
	// 1. Unify service names to lowercase for easier search
	span.ServiceName = strings.ToLower(span.ServiceName)

	// 2. Add default tags
	if span.Tags == nil {
		span.Tags = make(map[string]interface{})
	}
	span.Tags["ingested_at"] = time.Now().Format(time.RFC3339)

	// 3. Normalize timestamps to milliseconds if they appear to be in seconds
	// Assuming 1e12 as a threshold to distinguish seconds from milliseconds
	if span.StartTime < 1e12 { // Likely in seconds
		span.StartTime *= 1e3
	}
}
