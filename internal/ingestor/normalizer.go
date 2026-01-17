package ingestor

import (
	"strings"
	"time"
	"tracer/pkg/model"
)

type Normalizer struct {
}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) Normalize(span *model.FlatSpan) {
	// 1. 统一服务名为小写，方便检索
	span.ServiceName = strings.ToLower(span.ServiceName)

	// 2. 补全默认标签
	if span.Tags == nil {
		span.Tags = make(map[string]interface{})
	}
	span.Tags["ingested_at"] = time.Now().Format(time.RFC3339)

	// 3. 假设我们要将所有时间戳统一为毫秒秒 (ms)
	// 这里可以根据数值长度判断输入是秒还是毫秒，并强制转换
	if span.StartTime < 1e12 { // 可能是秒级时间戳
		span.StartTime *= 1e3
	}
}
