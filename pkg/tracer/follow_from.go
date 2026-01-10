package tracer

import "tracer/pkg/span"

type FollowFromOption struct {
	ctx span.SpanContext
}

func (o *FollowFromOption) Apply(s *StartSpanOption) {
	s.TracerID = o.ctx.TraceID
	s.References = append(s.References, span.Reference{
		RefType: span.FollowFrom,
		TraceID: o.ctx.TraceID,
		SpanID:  o.ctx.SpanID,
	})

	s.Baggage = o.ctx.Baggage
}

func FollowFrom(ctx span.SpanContext) *FollowFromOption {
	return &FollowFromOption{ctx: ctx}
}
