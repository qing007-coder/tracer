package sampler

import "tracer/pkg/config"

type Base struct {
	Type     string
	Param    float64
	Decision bool
	Tags     []config.Tag
}
