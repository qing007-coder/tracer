package span

import (
	"time"
	"tracer/pkg/config"
)

type Log struct {
	Timestamp time.Time
	Fields    []config.Tag
}

func String(key, value string) config.Tag {
	return config.Tag{
		Key:   key,
		Value: value,
	}
}

func Int(key string, value int) config.Tag {
	return config.Tag{
		Key:   key,
		Value: value,
	}
}
