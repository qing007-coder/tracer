package sampler

import "tracer/pkg/config"

type SamplerConst struct {
	Base
}

func NewSamplerConst(param float64) *SamplerConst {
	s := &SamplerConst{
		Base{
			Type:  SamplerTypeConst,
			Param: param,
			Tags: []config.Tag{
				{Key: SamplerTypeTagKey, Value: SamplerTypeConst},
				{Key: SamplerParamTagKey, Value: param},
			},
		},
	}

	s.init()
	return s
}

func (s *SamplerConst) init() {
	s.Decision = s.Param == 1
}

func (s *SamplerConst) IsSample(traceID, operation string) bool {
	return s.Decision
}
