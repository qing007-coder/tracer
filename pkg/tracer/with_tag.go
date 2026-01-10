package tracer

import "tracer/pkg/config"

type WithTagOption struct {
	key   string
	value interface{}
}

func (o *WithTagOption) Apply(s *StartSpanOption) {
	s.Tags = append(s.Tags, config.Tag{
		Key:   o.key,
		Value: o.value,
	})
}

func WithTag(key string, value interface{}) *WithTagOption {
	return &WithTagOption{
		key:   key,
		value: value,
	}
}
