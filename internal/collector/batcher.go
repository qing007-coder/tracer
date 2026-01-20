package collector

import (
	"fmt"
	"log"
	"sync"
	"time"
	"tracer/pkg/model"
)

// Batcher collects spans into batches before sending them to the dispatcher.
// It helps reduce the number of small messages sent to Kafka.
type Batcher struct {
	mu          sync.Mutex
	spanChan    <-chan *model.FlatSpan
	maxQueue    uint
	batchSpans  []*model.FlatSpan
	timer       *time.Timer
	duration    time.Duration
	outputChan  chan<- []*model.FlatSpan
	rateLimiter *RateLimiter
}

// NewBatcher creates a new Batcher.
func NewBatcher(maxQueue uint, spanChan <-chan *model.FlatSpan, outputChan chan<- []*model.FlatSpan, duration time.Duration) (*Batcher, error) {
	b := new(Batcher)
	b.init(maxQueue, spanChan, outputChan, duration)

	return b, nil
}

// init initializes the Batcher components.
func (b *Batcher) init(maxQueue uint, spanChan <-chan *model.FlatSpan, outputChan chan<- []*model.FlatSpan, duration time.Duration) {
	b.maxQueue = maxQueue
	b.spanChan = spanChan
	b.outputChan = outputChan
	b.duration = duration
	b.rateLimiter = NewRateLimiter()
	b.timer = time.NewTimer(duration)
	b.batchSpans = make([]*model.FlatSpan, 0, b.maxQueue)
}

// Start runs the batcher loop in a goroutine.
func (b *Batcher) Start() {
	go b.Run()
}

// Run listens for incoming spans and timer events to trigger batch sending.
func (b *Batcher) Run() {
	for {
		select {
		case span, ok := <-b.spanChan:
			if !ok {
				fmt.Println("batcher closed")
				b.Send()
				return
			}
			fmt.Println("span:", span)
			b.Append(span)
		case <-b.timer.C:
			b.Send()

			b.timer.Reset(b.duration)
		}
	}
}

//func (b *Batcher) Run() {
//	fmt.Println("Batcher Loop 确认启动...")
//	for span := range b.spanChan { // 直接 range 监听
//		fmt.Printf("Batcher 强行捕获数据: %+v\n", span)
//	}
//}

// Append adds a span to the current batch.
// It checks rate limits and sends the batch if it's full.
func (b *Batcher) Append(span *model.FlatSpan) {
	if !b.rateLimiter.Allow() {
		return
	}

	b.mu.Lock()
	b.batchSpans = append(b.batchSpans, span)
	isFull := len(b.batchSpans) == int(b.maxQueue)
	b.mu.Unlock()

	if isFull {
		b.Send()
	}
}

// Flush clears the current batch buffer.
func (b *Batcher) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.batchSpans = make([]*model.FlatSpan, 0, b.maxQueue)
}

// Send sends the current batch of spans to the output channel.
func (b *Batcher) Send() {
	b.mu.Lock()
	if len(b.batchSpans) == 0 {
		b.mu.Unlock()
		return
	}

	snapshot := b.batchSpans
	b.mu.Unlock()
	b.Flush()

	select {
	case b.outputChan <- snapshot:
		// 发送成功
	default:
		// 队列满了，直接丢弃这批数据并记录日志/指标
		// 这样能保证 batcher 永远不会卡死
		log.Println("output channel is full")
	}
}
