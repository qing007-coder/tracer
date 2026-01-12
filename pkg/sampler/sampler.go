package sampler

import "tracer/pkg/config"

type Sampler interface {
	IsSample(traceID, operation string) bool
	GetTags() []config.Tag
	init()
}

func NewSampler(conf *config.Configuration) Sampler {
	switch conf.Sampler.Type {
	case SamplerTypeConst:
		return NewSamplerConst(conf.Sampler.Param)
	case SamplerTypeProbabilistic:
	case SamplerTypeRateLimiting:
	}

	return nil
}
