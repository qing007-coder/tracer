package tracer

import (
	"context"
	"strconv"
	"strings"
	"time"
	"tracer/pkg/config"
	"tracer/pkg/model"
	"tracer/pkg/reporter"
	"tracer/pkg/sampler"
	"tracer/pkg/span"
	"tracer/pkg/utils"
)

type Tracer struct {
	ServiceName string
	Process     *model.Process
	Reporter    *reporter.Reporter
	Sampler     sampler.Sampler
}

func NewTracer(conf *config.Configuration, tags ...config.Tag) (*Tracer, error) {
	tracer := new(Tracer)
	err := tracer.init(conf, tags...)
	if err != nil {
		return nil, err
	}

	return tracer, nil
}

func (t *Tracer) init(conf *config.Configuration, tags ...config.Tag) error {
	t.ServiceName = conf.ServiceName
	r, err := reporter.NewReporter(conf, tags...)
	if err != nil {
		return err
	}
	t.Reporter = r
	t.Sampler = sampler.NewSampler(conf)
	t.Process = model.NewProcess(conf.ServiceName, t.Sampler.GetTags()...)
	t.Process.Tags = append(t.Process.Tags, tags...)
	t.Reporter.Start()

	return nil
}

func (t *Tracer) StartSpan(operation string, options ...Option) *span.Span {

	startSpanOption := new(StartSpanOption)

	for _, option := range options {
		option.Apply(startSpanOption)
	}

	var traceID string

	if startSpanOption.TracerID != "" {
		traceID = startSpanOption.TracerID
	} else {
		traceID = utils.CreateID()
	}

	baggage := make(map[string]string)

	if len(startSpanOption.Baggage) != 0 {
		baggage = startSpanOption.Baggage
	}

	return &span.Span{
		Operation: operation,
		Context: span.SpanContext{
			TraceID: traceID,
			SpanID:  utils.CreateID(),
			Sampled: t.Sampler.IsSample(traceID, operation),
			Baggage: baggage,
		},
		StartTime: time.Now(),
		ProcessID: utils.CreateID(),
		OnFinish: func(s *span.ToModel) {
			t.Reporter.Store(*s)
		},
		Tags:       startSpanOption.Tags,
		References: startSpanOption.References,
	}

}

// Inject 进程外 即跨服务用
func (t *Tracer) Inject(sc span.SpanContext, carrier Carrier) error {
	carrier.Set("tracer_id", sc.TraceID)
	carrier.Set("span_id", sc.SpanID)
	carrier.Set("sampled", strconv.FormatBool(sc.Sampled))

	sc.ForeachBaggageItem(func(k, v string) {
		carrier.Set("baggage_"+k, v)
	})

	return nil
}

// Extract 进程外 即跨服务用
func (t *Tracer) Extract(carrier Carrier) (span.SpanContext, error) {
	sc := span.NewSpanContext()

	carrier.Foreach(func(key string, v interface{}) {
		value := v.(string)

		switch key {
		case "tracer_id":
			sc.TraceID = value
		case "span_id":
			sc.SpanID = value
		case "sampled":
			sc.Sampled, _ = strconv.ParseBool(value)
		default:
			if strings.HasPrefix(key, "baggage_") {
				realKey := strings.TrimPrefix(key, "baggage_")
				sc.Baggage[realKey] = value
			}
		}

	})

	//sc.TraceID = carrier.Get("tracer_id").(string)
	//sc.ParentID = carrier.Get("span_id").(string)
	//sc.Sampled = carrier.Get("sampled").(bool)
	//
	//sc.ForeachBaggageItem(func(k, v string) {
	//	sc.Baggage[k] = carrier.Get("baggage_" + k).(string)
	//})

	return sc, nil
}

type SpanKeyType struct {
}

var spanKey = SpanKeyType{}

// SpanFromContext 进程内部使用（即服务内部）
func (t *Tracer) SpanFromContext(ctx context.Context) *span.Span {
	return ctx.Value(spanKey).(*span.Span)
}

// ContextFromSpan 进程内部使用（即服务内部）
func (t *Tracer) ContextFromSpan(ctx context.Context, span *span.Span) context.Context {
	return context.WithValue(ctx, spanKey, span)
}
