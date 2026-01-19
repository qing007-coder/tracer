package utils

import (
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	pb "tracer/internal/proto"
	"tracer/pkg/model"
)

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func mapToStringMap(in map[string]interface{}) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = toString(v)
	}
	return out
}

func timestampToMicro(ts *timestamppb.Timestamp) int64 {
	if ts == nil {
		return 0
	}
	return ts.Seconds*1_000_000 + int64(ts.Nanos)/1_000
}

func flattenLogs(
	logs []*pb.Log,
) (names []string, times []int64, attrs []map[string]string) {

	for _, log := range logs {
		if log == nil {
			continue
		}

		names = append(names, "log")
		times = append(times, timestampToMicro(log.Timestamp))

		// fields 已经是 map[string]string，直接用
		if len(log.Fields) > 0 {
			attrs = append(attrs, log.Fields)
		} else {
			attrs = append(attrs, nil)
		}
	}

	return
}

func FlatSpanToClickHouseSpan(fs *model.FlatSpan) *model.StorageSpan {
	if fs == nil {
		return nil
	}

	// Tags = Span tags + Process tags（合并）
	attrs := make(map[string]string)

	for k, v := range fs.Tags {
		attrs[k] = toString(v)
	}
	for k, v := range fs.ProcessTags {
		attrs["process."+k] = toString(v)
	}

	// Logs → Events
	eventNames, eventTimes, eventAttrs := flattenLogs(fs.Logs)
	fmt.Println("parent_id", fs.ParentSpanID)

	return &model.StorageSpan{
		TraceID:      fs.TraceID,
		SpanID:       fs.SpanID,
		ParentSpanID: fs.ParentSpanID,

		ServiceName:   fs.ServiceName,
		OperationName: fs.OperationName,
		Kind:          inferSpanKind(fs.Tags), // 可选

		StartTimeUs: fs.StartTime,
		DurationUs:  fs.Duration,

		StatusCode:    inferStatus(fs.Tags),
		StatusMessage: "",

		Attributes: attrs,

		EventNames:   eventNames,
		EventTimesUs: eventTimes,
		EventAttrs:   eventAttrs,

		ProcessID:     fs.ProcessID,
		ResourceAttrs: mapToStringMap(fs.ProcessTags),

		TimestampUs: fs.StartTime,
	}
}

func inferSpanKind(tags map[string]interface{}) string {
	if v, ok := tags["span.kind"]; ok {
		return strings.ToUpper(toString(v))
	}
	return "INTERNAL"
}

func inferStatus(tags map[string]interface{}) string {
	if v, ok := tags["error"]; ok {
		if toString(v) == "true" || toString(v) == "1" {
			return "ERROR"
		}
	}
	return "OK"
}
