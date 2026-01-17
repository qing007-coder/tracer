package model

// StorageSpan 一行 = 一个 Span
type StorageSpan struct {
	// Trace 关系
	TraceID      string
	SpanID       string
	ParentSpanID string

	// 基本信息
	ServiceName   string
	OperationName string
	Kind          string // SERVER / CLIENT / INTERNAL

	// 时间（统一用微秒或纳秒，下面用微秒）
	StartTimeUs int64
	DurationUs  int64

	// 状态
	StatusCode    string // OK / ERROR / UNSET
	StatusMessage string

	// Tags / Attributes
	Attributes map[string]string

	// Events / Logs（拆列）
	EventNames   []string
	EventTimesUs []int64
	EventAttrs   []map[string]string

	// Process / Resource
	ProcessID     string
	ResourceAttrs map[string]string

	// 采样 & 其他
	Sampled bool
	Baggage map[string]string

	// 用于分区 & TTL
	TimestampUs int64
}
