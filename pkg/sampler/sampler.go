package sampler

type Sampler interface {
	IsSample(id, operation string) bool
}
