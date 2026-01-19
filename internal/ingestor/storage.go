package ingestor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
	"tracer/pkg/model"
)

type Storage struct {
	ctx          context.Context
	mu           sync.Mutex
	sampler      *TailSampler
	conn         clickhouse.Conn
	batchData    []*model.StorageSpan
	batchSize    int
	spanChan     chan []*model.StorageSpan
	timer        *time.Timer
	duration     time.Duration
	workerNum    int
	snapshotChan chan []*model.StorageSpan
}

func NewStorage(conf Configuration) *Storage {
	s := new(Storage)
	s.init(conf)

	return s
}

func (s *Storage) init(conf Configuration) {
	s.batchData = make([]*model.StorageSpan, 0)
	s.sampler = NewTailSampler(1, time.Second*2)
	s.batchSize = conf.BatchSize
	s.spanChan = conf.MergerToStorageChan
	s.ctx = context.Background()
	s.duration = time.Second * 3
	s.workerNum = 2
	s.snapshotChan = make(chan []*model.StorageSpan, 100)
	s.timer = time.NewTimer(s.duration)

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"192.168.233.128:9000"},
		Auth: clickhouse.Auth{
			Database: "tracer",
			Username: "default",
			Password: "",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.Ping(s.ctx); err != nil {
		log.Fatal("clickhouse ping failed:", err)
	}
	log.Println("clickhouse connected")

	s.conn = conn
}

func (s *Storage) Run() {
	go s.retryWorker(s.ctx)
	for i := 0; i < s.workerNum; i++ {
		go s.workerRun()
	}
	for {
		select {
		case <-s.timer.C:
			s.Flush()
			s.timer.Reset(s.duration)
		case spans := <-s.spanChan:
			if s.Store(spans) {
				if !s.timer.Stop() {
					select {
					case <-s.timer.C: // 避免 timer 已经触发但还没被消费的情况
					default:
					}
				}

				s.timer.Reset(s.duration)
			}
		}
	}
}

func (s *Storage) Store(spans []*model.StorageSpan) bool {
	if !s.sampler.IsSampled(spans) || len(spans) == 0 {
		return false
	}

	s.mu.Lock()
	s.batchData = append(s.batchData, spans...)
	isFull := len(s.batchData) >= s.batchSize

	s.mu.Unlock()
	if isFull {
		s.Flush()
		return true
	}

	return false
}

func (s *Storage) Flush() {
	s.mu.Lock()
	if len(s.batchData) == 0 {
		s.mu.Unlock()
		return
	}
	snapshot := s.batchData
	s.batchData = make([]*model.StorageSpan, 0, s.batchSize)
	s.mu.Unlock() // 必须先解锁

	// 此时即使 snapshotChan 满了，也只会阻塞当前 Flush 协程，不会锁住整个 Storage
	s.snapshotChan <- snapshot
}

func (s *Storage) WriteInDatabase(data []*model.StorageSpan) error {
	batch, err := s.conn.PrepareBatch(s.ctx, `
        INSERT INTO trace_spans (
            TraceID, SpanID, ParentSpanID,
            ServiceName, OperationName, Kind,
            StartTimeUs, DurationUs,
            StatusCode, StatusMessage,
            Attributes,
            EventNames, EventTimesUs, EventAttrs,
            ProcessID, ResourceAttrs,
            TimestampUs
        )
    `)
	if err != nil {
		log.Printf("PrepareBatch Error: %v", err)
		return err
	}

	// 3. 遍历快照，填充数据
	for _, span := range data {
		// 特殊处理：EventAttrs []map[string]string 转为 []string(JSON)
		// 因为 ClickHouse 的 Array(Map) 性能较差且驱动支持复杂
		eventAttrsStrs := make([]string, len(span.EventAttrs))
		for i, attrMap := range span.EventAttrs {
			if attrMap != nil {
				data, _ := json.Marshal(attrMap)
				eventAttrsStrs[i] = string(data)
			} else {
				eventAttrsStrs[i] = "{}"
			}
		}

		// 严格按照上面 INSERT 语句的顺序 Append
		err := batch.Append(
			span.TraceID,
			span.SpanID,
			span.ParentSpanID,
			span.ServiceName,
			span.OperationName,
			span.Kind,
			span.StartTimeUs,
			span.DurationUs,
			span.StatusCode,
			span.StatusMessage,
			span.Attributes,   // clickhouse-go 自动支持 map[string]string
			span.EventNames,   // []string
			span.EventTimesUs, // []int64
			eventAttrsStrs,    // []string (JSON 序列化后的数组)
			span.ProcessID,
			span.ResourceAttrs, // map[string]string
			span.TimestampUs,
		)

		if err != nil {
			log.Printf("Append Batch Error: %v, TraceID: %s", err, span.TraceID)
			continue
		}
	}

	// 4. 正式发送到 ClickHouse
	if err := batch.Send(); err != nil {
		log.Printf("Flush to ClickHouse Failed: %v", err)
		s.WriteInFile(data)
		return err
	}

	log.Printf("Successfully flushed %d spans to ClickHouse", len(data))
	return nil
}

func (s *Storage) WriteInFile(data []*model.StorageSpan) {
	if len(data) == 0 {
		return
	}

	// 增加：确保目录存在
	dir := "./data/"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create dir %s: %v", dir, err)
		return
	}

	fileName := fmt.Sprintf("spans-%s.jsonl", time.Now().Format("20060102-150405.000"))
	filePath := filepath.Join(dir, fileName)

	// 2. 打开文件（只写、创建、追加模式）
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open local file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	// 3. 使用带缓冲的写入器，减少系统调用次数
	writer := bufio.NewWriterSize(file, 1024*1024) // 1MB 缓冲区
	defer writer.Flush()

	encoder := json.NewEncoder(writer)

	// 4. 循环写入每条数据
	for _, span := range data {
		if err := encoder.Encode(span); err != nil {
			log.Printf("Failed to encode span to JSON: %v", err)
			continue
		}
	}

	log.Printf("Successfully wrote %d spans to local file: %s", len(data), filePath)
}

func (s *Storage) retryWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 开始扫表
			files, _ := filepath.Glob(filepath.Join("./data/", "*.jsonl"))
			for _, f := range files {
				// 1. 先改名，占坑，防止并发读写冲突
				processingName := f + ".processing"
				_ = os.Rename(f, processingName)

				// 2. 读取并尝试写入
				err := s.loadAndInsert(processingName)
				if err == nil {
					// 3. 只有成功了才删掉
					_ = os.Remove(processingName)
				} else {
					// 4. 如果还是写不进去（比如CK还没好），把名字改回来，等下次 ticker
					_ = os.Rename(processingName, f)
					break // 既然连这个都写不进去，后面的文件肯定也写不进去，直接跳出循环
				}
			}
		}
	}
}

func (s *Storage) loadAndInsert(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	defer file.Close()

	var recoveryBatch []*model.StorageSpan
	scanner := bufio.NewScanner(file)
	// 同样设置缓冲区，防止大 Tags 导致读取失败
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		var span model.StorageSpan
		if err := json.Unmarshal(scanner.Bytes(), &span); err != nil {
			log.Printf("Unmarshal error in file %s: %v", filePath, err)
			continue
		}
		recoveryBatch = append(recoveryBatch, &span)

		// 如果文件特别大（比如几十万条），分批写入 ClickHouse 更好
		if len(recoveryBatch) >= s.batchSize {
			if err := s.WriteInDatabase(recoveryBatch); err != nil {
				return fmt.Errorf("batch insert from file failed: %w", err)
			}
			recoveryBatch = recoveryBatch[:0] // 清空切片继续读
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	// 处理文件中剩余的（不足 batchSize 的部分）
	if len(recoveryBatch) > 0 {
		if err := s.WriteInDatabase(recoveryBatch); err != nil {
			return fmt.Errorf("final batch insert from file failed: %w", err)
		}
	}

	log.Printf("Successfully restored file: %s", filePath)
	return nil
}

func (s *Storage) workerRun() {
	for {
		select {
		case snapshot := <-s.snapshotChan:
			if err := s.WriteInDatabase(snapshot); err != nil {
				fmt.Println("err:", err)
			}
		}
	}

}
