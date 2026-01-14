package agent

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
	pb "tracer/internal/proto"
	"tracer/pkg/config"
	"tracer/pkg/model"
	"tracer/pkg/span"
)

type Exporter struct {
	ctx       context.Context
	WorkerNum int
	batchCh   <-chan model.BatchPackage
	client    pb.CollectorServiceClient
}

func NewExporter(workNum int, batchCh <-chan model.BatchPackage) (*Exporter, error) {
	e := new(Exporter)
	if err := e.init(workNum, batchCh); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Exporter) init(workNum int, batchCh <-chan model.BatchPackage) error {
	e.ctx = context.Background()
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}

	e.client = pb.NewCollectorServiceClient(conn)
	e.WorkerNum = workNum
	e.batchCh = batchCh

	return nil
}

func (e *Exporter) Start() {
	for i := 0; i < e.WorkerNum; i++ {
		go e.Consume()
	}
}

func (e *Exporter) Consume() {
	for {
		select {
		case <-e.ctx.Done():
			log.Println("exit")
			return
		case batch := <-e.batchCh:
			ctx, cancel := context.WithTimeout(e.ctx, 5*time.Second)
			resp, err := e.client.Export(ctx, e.BatchToModel(batch))
			if err != nil {
				log.Println(err)
			}

			cancel()

			log.Println(resp)
		}
	}

}

// BatchToModel 将业务模型转换为 PB 模型
func (e *Exporter) BatchToModel(bp model.BatchPackage) *pb.BatchPackage {
	if len(bp.Packages) == 0 {
		return &pb.BatchPackage{}
	}

	req := &pb.BatchPackage{
		Packages: make([]*pb.Package, 0, len(bp.Packages)),
	}

	for _, p := range bp.Packages {
		req.Packages = append(req.Packages, &pb.Package{
			Process: convertProcess(p.Process),
			Spans:   convertSpans(p.Spans),
		})
	}
	fmt.Println(req)

	return req
}

// 转换 Process
func convertProcess(p model.Process) *pb.Process {
	return &pb.Process{
		ServiceName: p.ServiceName,
		// 如果你的 pb.Process 定义了 ProcessId 字段：
		// ProcessId: p.ProcessID,
		Tags: tagsToMap(p.Tags),
	}
}

// 转换 Spans
func convertSpans(spans []span.ToModel) []*pb.SpanModel {
	res := make([]*pb.SpanModel, 0, len(spans))
	for _, s := range spans {
		sm := &pb.SpanModel{
			Operation: s.Operation,
			Context: &pb.SpanContext{
				TraceId: s.Context.TraceID,
				SpanId:  s.Context.SpanID,
			},
			Tags:      tagsToMap(s.Tags),
			StartTime: timestamppb.New(s.StartTime),
			Duration:  durationpb.New(s.Duration),
		}

		// 转换 References
		for _, ref := range s.References {
			sm.References = append(sm.References, &pb.Reference{
				TraceId: ref.TraceID,
				SpanId:  ref.SpanID,
				RefType: ref.RefType,
			})
		}

		// 转换 Logs
		for _, l := range s.Logs {
			sm.Logs = append(sm.Logs, &pb.Log{
				Timestamp: timestamppb.New(l.Timestamp),
				Fields:    tagsToMap(l.Fields),
			})
		}

		res = append(res, sm)
	}
	return res
}

// 核心转换函数：处理 interface{} 类型的 Tag
func tagsToMap(tags []config.Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}

	res := make(map[string]string, len(tags))
	for _, t := range tags {
		// 使用 fmt.Sprintf 确保无论 interface{} 是 int, bool 还是 string
		// 都能安全地转为字符串，防止报错
		res[t.Key] = fmt.Sprintf("%v", t.Value)
	}
	return res
}
