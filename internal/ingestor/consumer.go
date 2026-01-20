package ingestor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"hash/fnv"
	"log"
	"tracer/pkg/model"
)

// Consumer consumes messages from Kafka.
type Consumer struct {
	workerNum  int
	group      sarama.ConsumerGroup
	workerChan []chan *model.FlatSpan
	validator  *Validator
	normalizer *Normalizer
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(conf Configuration) (*Consumer, error) {
	c := new(Consumer)
	err := c.init(conf)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// init initializes the Consumer with Sarama configuration and worker channels.
func (c *Consumer) init(conf Configuration) error {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	group, err := sarama.NewConsumerGroup([]string{"192.168.233.128:9092"}, "tracer-group-receiver", config)
	if err != nil {
		return err
	}

	c.workerNum = conf.WorkerNum
	c.workerChan = conf.ConsumerToMergerChan

	c.group = group
	c.normalizer = NewNormalizer()
	c.validator = NewValidator()
	return nil
}

// Run starts the consumer loop.
func (c *Consumer) Run() {
	ctx := context.Background()

	for {
		if err := c.group.Consume(ctx, []string{"span-tracer-test"}, c); err != nil {
			log.Printf("consume error: %v", err)
		}

		if ctx.Err() != nil {
			return
		}
	}
}

// HandleMessage processes a single Kafka message: unmarshal, validate, normalize, and route to worker.
func (c *Consumer) HandleMessage(message *sarama.ConsumerMessage) {
	span := new(model.FlatSpan)
	if err := json.Unmarshal(message.Value, span); err != nil {
		fmt.Println("err:", err)
		return
	}
	if err := c.validator.Validate(span); err != nil {
		fmt.Println("err:", err)
		return
	}

	c.normalizer.Normalize(span)

	id := c.route(span.TraceID)
	fmt.Println(span)

	c.workerChan[id] <- span
	//select {
	//case c.workerChan[id] <- span:
	//	fmt.Println("success")
	//default:
	//	fmt.Println("out")
	//}

}

// route calculates the worker ID based on TraceID to ensure same trace goes to same worker.
func (c *Consumer) route(traceID string) int {
	h := fnv.New32a()
	_, err := h.Write([]byte(traceID))
	if err != nil {
		return 0
	}
	return int(h.Sum32()) % c.workerNum
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.HandleMessage(msg)
		session.MarkMessage(msg, "")
	}
	return nil
}
