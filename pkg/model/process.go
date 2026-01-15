package model

import (
	"tracer/pkg/config"
)

type Process struct {
	ID          string
	ServiceName string
	Tags        []config.Tag
}

func NewProcess(serviceName string, tags ...config.Tag) *Process {
	return &Process{
		ServiceName: serviceName,
		Tags:        tags,
	}
}
