package tracer

import "tracer/pkg/span"

type ChildOfOption struct {
	ctx span.SpanContext
}

func (o *ChildOfOption) Apply(s *StartSpanOption) {
	s.TracerID = o.ctx.TraceID
	s.References = append(s.References, span.Reference{
		RefType: span.ChildOf,
		TraceID: o.ctx.TraceID,
		SpanID:  o.ctx.SpanID,
	})

	s.Baggage = o.ctx.Baggage
}

func ChildOf(ctx span.SpanContext) *ChildOfOption {
	return &ChildOfOption{ctx: ctx}
}
