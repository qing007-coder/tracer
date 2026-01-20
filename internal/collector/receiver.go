package collector

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"google.golang.org/grpc"
	"log"
	"net"
	"sort"
	pb "tracer/internal/proto"
	"tracer/pkg/model"
)

// Receiver handles incoming gRPC requests from Agents.
type Receiver struct {
	pb.UnimplementedCollectorServiceServer
	server    *grpc.Server
	lis       net.Listener
	spanChan  chan<- *model.FlatSpan
	validator *Validator
}

// NewReceiver creates a new Receiver.
func NewReceiver(spanCh chan<- *model.FlatSpan) (*Receiver, error) {
	r := new(Receiver)
	if err := r.init(spanCh); err != nil {
		return nil, err
	}

	return r, nil
}

// init initializes the gRPC server and validator.
func (r *Receiver) init(spanCh chan<- *model.FlatSpan) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 2. 实例化 gRPC 服务端
	r.server = grpc.NewServer()
	pb.RegisterCollectorServiceServer(r.server, r)

	r.lis = lis
	r.spanChan = spanCh
	v, err := NewValidator()
	if err != nil {
		return err
	}

	r.validator = v
	return nil
}

// Run starts the gRPC server.
func (r *Receiver) Run() {
	log.Println("gRPC 服务端启动在 :50051...")
	if err := r.server.Serve(r.lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Export implements the gRPC Export method to receive trace data.
// It validates the batch and pushes flattened spans to the span channel.
func (r *Receiver) Export(ctx context.Context, batchPkg *pb.BatchPackage) (*pb.ExportResponse, error) {
	if err := r.validator.Validate(batchPkg); err != nil {
		return nil, err
	}

	spans := r.ConvertBatchToFlatSpans(batchPkg)

	for _, span := range spans {
		r.spanChan <- span
	}

	return &pb.ExportResponse{
		Success: true,
	}, nil
}

// ConvertBatchToFlatSpans converts a protobuf BatchPackage to a slice of FlatSpans.
func (r *Receiver) ConvertBatchToFlatSpans(batchPkg *pb.BatchPackage) []*model.FlatSpan {
	var allFlatSpans []*model.FlatSpan

	for _, pkg := range batchPkg.GetPackages() {
		process := pkg.GetProcess()
		if process == nil {
			continue
		}

		pID := process.GetProcessId()
		if pID == "" {
			pID = GenerateProcessID(process)
		}

		pTags := convertTags(process.GetTags())

		for _, s := range pkg.GetSpans() {
			allFlatSpans = append(allFlatSpans, &model.FlatSpan{
				TraceID:       s.GetContext().GetTraceId(),
				SpanID:        s.GetContext().GetSpanId(),
				ParentSpanID:  s.GetContext().GetParentId(),
				OperationName: s.GetOperation(),
				Baggage:       convertTags(s.GetContext().GetBaggage()),
				Sampled:       s.GetContext().GetSampled(),
				StartTime:     s.GetStartTime().AsTime().UnixMicro(),
				Duration:      s.GetDuration().AsDuration().Microseconds(),
				Tags:          convertTags(s.GetTags()),

				Logs:       s.GetLogs(),
				References: s.GetReferences(),

				ProcessID:   pID,
				ServiceName: process.GetServiceName(),
				ProcessTags: pTags,
			})
		}
	}
	return allFlatSpans
}

// GenerateProcessID 辅助函数
func GenerateProcessID(p *pb.Process) string {
	h := md5.New()
	h.Write([]byte(p.GetServiceName()))
	keys := make([]string, 0, len(p.GetTags()))
	for k := range p.GetTags() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(p.GetTags()[k]))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// convertTags 转换 Map
func convertTags(tags map[string]string) map[string]interface{} {
	res := make(map[string]interface{}, len(tags))
	for k, v := range tags {
		res[k] = v
	}
	return res
}
