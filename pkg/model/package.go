package model

import (
	pb "tracer/internal/proto"
	"tracer/pkg/span"
)

type Package struct {
	Process Process
	Spans   []span.ToModel
}

type BatchPackage struct {
	Packages []Package
}

type FlatSpan struct {
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	ParentSpanID  string                 `json:"parent_id"` // 方便直接查父节点
	Sampled       bool                   `json:"sampled"`
	Baggage       map[string]interface{} `json:"baggage"`
	OperationName string                 `json:"operation"`
	StartTime     int64                  `json:"start_time"` // 建议统一用微秒，方便计算
	Duration      int64                  `json:"duration"`   // 微秒
	Tags          map[string]interface{} `json:"tags"`       // 把 []config.Tag 转成 Map，方便后端索引
	Logs          []*pb.Log              `json:"logs"`
	// 3. 关联关系
	References []*pb.Reference `json:"references"`

	// 4. 扁平化注入的 Process 信息 (这是重点！)
	ProcessID   string                 `json:"process_id"`   // 你搓出来的 Hash ID
	ServiceName string                 `json:"service_name"` // 顶层索引字段
	ProcessTags map[string]interface{} `json:"process_tags"` // 比如 IP, Hostname 等
}
