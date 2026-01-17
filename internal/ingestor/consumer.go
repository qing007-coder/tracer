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

type Consumer struct {
	workerNum  int
	group      sarama.ConsumerGroup
	workerChan []chan *model.FlatSpan
	validator  *Validator
	normalizer *Normalizer
}

func NewConsumer(conf Configuration) (*Consumer, error) {
	c := new(Consumer)
	err := c.init(conf)
	if err != nil {
		return nil, err
	}

	return c, nil
}

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

func (c *Consumer) route(traceID string) int {
	h := fnv.New32a()
	_, err := h.Write([]byte(traceID))
	if err != nil {
		return 0
	}
	return int(h.Sum32()) % c.workerNum
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.HandleMessage(msg)
		session.MarkMessage(msg, "")
	}
	return nil
}
