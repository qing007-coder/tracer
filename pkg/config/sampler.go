package config

type SamplerConfig struct {
	Type  string  `json:"type"`
	Param float64 `json:"param"`
}

func NewSampler() {

}

func (c *SamplerConfig) name() {

}
