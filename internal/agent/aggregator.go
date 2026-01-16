package agent

import (
	"log"
	"sync"
	"time"
	"tracer/pkg/model"
	"tracer/pkg/span"
)

type Aggregator struct {
	mu           sync.Mutex
	batchCh      <-chan model.Package
	maxQueue     uint
	batchPackage map[string]*model.Package
	timer        *time.Timer
	outputCh     chan<- model.BatchPackage
	duration     time.Duration
}

func NewAggregator(
	duration time.Duration,
	batchCh <-chan model.Package,
	outputCh chan<- model.BatchPackage,
	maxQueue uint,
) *Aggregator {
	a := &Aggregator{
		duration:     duration,
		timer:        time.NewTimer(duration),
		batchPackage: make(map[string]*model.Package),
		outputCh:     outputCh,
		maxQueue:     maxQueue,
		batchCh:      batchCh,
	}
	return a
}

func (a *Aggregator) Start() {
	go a.Run()
}

func (a *Aggregator) Run() {
	for {
		select {
		case batch := <-a.batchCh:
			if a.Append(batch) {
				if !a.timer.Stop() {
					select {
					case <-a.timer.C:
					default:

					}
				}

				a.timer.Reset(a.duration)
			}

		case <-a.timer.C:
			a.sendAll()
			a.timer.Reset(a.duration)
		}
	}
}

func (a *Aggregator) Append(pkg model.Package) bool {
	service := pkg.Process.ServiceName

	a.mu.Lock()

	// 第一次出现该 service，初始化
	bp, ok := a.batchPackage[service]
	if !ok {
		bp = &model.Package{
			Process: pkg.Process,
		}
		a.batchPackage[service] = bp
	}

	bp.Spans = append(bp.Spans, pkg.Spans...)
	isFull := len(bp.Spans) >= int(a.maxQueue)

	a.mu.Unlock()

	if isFull {
		a.Send(service)
		a.Flush(service)
		return true
	}

	return false
}

func (a *Aggregator) Send(serviceName string) {
	a.mu.Lock()
	pkg, ok := a.batchPackage[serviceName]
	if !ok || len(pkg.Spans) == 0 {
		a.mu.Unlock()
		return
	}

	out := model.Package{
		Process: pkg.Process,
		Spans:   append([]span.ToModel(nil), pkg.Spans...),
	}
	a.mu.Unlock()

	bp := model.BatchPackage{
		Packages: []model.Package{
			out,
		},
	}

	select {
	case a.outputCh <- bp:
		// 发送成功
	default:
		// 队列满了，直接丢弃这批数据并记录日志/指标
		// 这样能保证 Agent 永远不会卡死
		log.Println("Aggregator output channel full, dropping batch")
	}
}

func (a *Aggregator) Flush(serviceName string) {
	a.mu.Lock()
	delete(a.batchPackage, serviceName)
	a.mu.Unlock()
}

func (a *Aggregator) sendAll() {
	a.mu.Lock()
	snapshot := a.batchPackage
	a.batchPackage = make(map[string]*model.Package)
	a.mu.Unlock()

	if len(snapshot) == 0 {
		return
	}

	var bp model.BatchPackage
	for _, pkg := range snapshot {
		if len(pkg.Spans) == 0 {
			continue
		}

		bp.Packages = append(bp.Packages, *pkg)
	}

	a.outputCh <- bp
}
