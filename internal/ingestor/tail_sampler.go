package ingestor

import (
	"math/rand"
	"time"
	"tracer/pkg/model"
)

// TailSampler determines whether a trace should be stored based on its characteristics (error, latency) and a sample rate.
type TailSampler struct {
	sampleRate    float64
	slowThreshold time.Duration
}

// NewTailSampler creates a new TailSampler.
func NewTailSampler(sampleRate float64, slowThreshold time.Duration) *TailSampler {
	return &TailSampler{
		sampleRate:    sampleRate,
		slowThreshold: slowThreshold,
	}
}

// IsSampled checks if the trace (collection of spans) should be sampled.
// It prioritizes traces with errors or slow requests.
func (t *TailSampler) IsSampled(spans []*model.StorageSpan) bool {
	hasError := false
	isSlow := false

	for _, span := range spans {
		// Check for errors
		if span.StatusCode == "ERROR" {
			hasError = true
			break
		}
		// Check for slow calls
		if span.DurationUs > t.slowThreshold.Microseconds() {
			isSlow = true
		}
	}

	// Always sample if there's an error or it's a slow request
	if hasError || isSlow {
		return true
	}

	// Otherwise sample probabilistically
	return rand.Float64() < t.sampleRate
}
