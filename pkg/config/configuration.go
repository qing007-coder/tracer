package config

type Configuration struct {
	ServiceName string
	Sampler     *SamplerConfig
	Reporter    *ReporterConfig
}
