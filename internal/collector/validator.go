package collector

import (
	"fmt"
	"time"
	pb "tracer/internal/proto"
)

type Validator struct {
}

func NewValidator() (*Validator, error) {
	return &Validator{}, nil
}

func (v *Validator) Validate(pkg *pb.BatchPackage) error {
	for _, pkg := range pkg.GetPackages() {
		// 1. 校验 Process
		if pkg.GetProcess().GetServiceName() == "" {
			return fmt.Errorf("missing service name")
		}

		// 2. 校验 Spans
		for _, s := range pkg.GetSpans() {
			ctx := s.GetContext()
			if ctx.GetTraceId() == "" || ctx.GetSpanId() == "" {
				return fmt.Errorf("missing trace/span id")
			}

			// 3. 时间合法性 (修正或拦截)
			startTime := s.GetStartTime().AsTime()
			if time.Since(startTime) > 24*time.Hour {
				return fmt.Errorf("span is too old")
			}

			// 4. 长度限制 (防止大字段挤爆 Kafka)
			for k, v := range s.GetTags() {
				if len(k) > 128 || len(v) > 2048 {
					return fmt.Errorf("tag too large")
				}
			}
		}
	}
	return nil
}
