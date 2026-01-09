package transport

import (
	"encoding/json"
	"sync"
	"tracer/pkg/config"
	"tracer/pkg/model"
	"tracer/pkg/span"
	"tracer/pkg/tracer"
)

type Batch struct {
	mu       sync.Mutex
	maxQueue uint
	process  *tracer.Process
	spans    []span.Span
	fullChan chan struct{}
}

func NewBatch(conf *config.Configuration, fullChan chan struct{}, tags ...config.Tag) *Batch {
	batch := new(Batch)
	batch.init(conf, fullChan, tags...)

	return batch
}

func (b *Batch) init(conf *config.Configuration, fullCh chan struct{}, tags ...config.Tag) {
	b.maxQueue = conf.Reporter.QueueSize
	b.process = tracer.NewProcess(conf.ServiceName, tags...)
	b.spans = make([]span.Span, 0, b.maxQueue)
	b.fullChan = fullCh
}

func (b *Batch) Start() {}

func (b *Batch) Flush() {

	b.spans = b.spans[:0] // 不加锁是因为用这个函数的时候已经锁了
}

func (b *Batch) Push(span span.Span) {
	b.mu.Lock()
	b.spans = append(b.spans, span)
	isFull := len(b.spans) >= cap(b.spans)
	b.mu.Unlock()

	if isFull {
		select {
		case b.fullChan <- struct{}{}: // 成功发送信号，Reporter 将开始工作
		default:
			// 信号发送失败（Reporter 还没处理完上一次信号）
			// 我们什么都不做。因为：
			// 1. 数据已在 b.spans 里，下次 Reporter 动作时会一并带走
			// 2. 避免了因为 Reporter 慢而卡死业务主逻辑
		}
	}
}

func (b *Batch) GetData() ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p := model.Package{
		Process: *b.process,
		Spans:   b.spans,
	}

	data, err := json.Marshal(&p)

	b.Flush()

	return data, err
}
