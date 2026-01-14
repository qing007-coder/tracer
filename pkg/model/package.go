package model

import "tracer/pkg/span"

type Package struct {
	Process Process
	Spans   []span.ToModel
}

type BatchPackage struct {
	Packages []Package
}
