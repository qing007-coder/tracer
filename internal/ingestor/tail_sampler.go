package ingestor

import (
	"math/rand"
	"time"
	"tracer/pkg/model"
)

type TailSampler struct {
	sampleRate    float64
	slowThreshold time.Duration
}

func NewTailSampler(sampleRate float64, slowThreshold time.Duration) *TailSampler {
	return &TailSampler{
		sampleRate:    sampleRate,
		slowThreshold: slowThreshold,
	}
}

func (t *TailSampler) IsSampled(spans []*model.StorageSpan) bool {
	hasError := false
	isSlow := false

	for _, span := range spans {
		// 校验是否报错
		if span.StatusCode == "ERROR" {
			hasError = true
			break
		}
		// 校验是否是慢调用
		if span.DurationUs > t.slowThreshold.Microseconds() {
			isSlow = true
		}
	}

	// 只要有错或者是慢请求，100% 存储
	if hasError || isSlow {
		return true
	}

	// 剩下的正常请求，按比例抽样
	return rand.Float64() < t.sampleRate
}
