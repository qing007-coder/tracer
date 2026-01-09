package sampler

import "tracer/pkg/config"

type Sampler interface {
	IsSample(traceID, operation string) bool
	init()
}

func NewSampler(conf *config.Configuration) Sampler {
	switch conf.Sample.Type {
	case SamplerTypeConst:
		return NewSamplerConst(conf.Sample.Param)
	case SamplerTypeProbabilistic:
	case SamplerTypeRateLimiting:
	}

	return nil
}
