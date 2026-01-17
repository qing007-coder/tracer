package collector

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"time"
	"tracer/pkg/model"
)

type Dispatcher struct {
	workerNum int
	producer  sarama.AsyncProducer
	batchChan <-chan []*model.FlatSpan
}

func NewDispatcher(workNum int, batchChan <-chan []*model.FlatSpan) (*Dispatcher, error) {
	d := new(Dispatcher)
	if err := d.init(workNum, batchChan); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Dispatcher) init(workerNum int, batchChan <-chan []*model.FlatSpan) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Flush.Messages = 500                     // 队列满 500 条再发
	config.Producer.Flush.Frequency = 100 * time.Millisecond // 或者每 100ms 发一次
	config.Producer.MaxMessageBytes = 1000000                // 单条消息最大字节数

	p, err := sarama.NewAsyncProducer([]string{"192.168.233.128:9092"}, config)
	if err != nil {
		return err
	}

	d.producer = p
	d.batchChan = batchChan
	d.workerNum = workerNum

	return nil
}

func (d *Dispatcher) Start() {
	go d.listen()
	for i := 0; i < d.workerNum; i++ {
		go d.Run()
	}
}

func (d *Dispatcher) Run() {
	for {
		select {
		case spans, ok := <-d.batchChan:
			if !ok {
				fmt.Println("out")
				return
			}
			for _, span := range spans {
				d.Send("span-tracer-test", span.TraceID, span)
			}

		}
	}
}

func (d *Dispatcher) Send(topic string, key string, value *model.FlatSpan) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Println(err)
		return
	}

	d.producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}
}

func (d *Dispatcher) listen() {
	for {
		select {
		case err, ok := <-d.producer.Errors():
			if !ok {
				return
			}
			log.Println("err:", err.Error())
		case success, ok := <-d.producer.Successes():
			if !ok {
				return
			}
			log.Println("success:", success.Key)
		}
	}
}

func (d *Dispatcher) Close() error {
	return d.producer.Close()
}
