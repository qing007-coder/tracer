package reporter

import (
	"fmt"
	"net"
	"time"
	"tracer/pkg/config"
	"tracer/pkg/span"
	"tracer/pkg/transport"
)

type Reporter struct {
	addr     *net.UDPAddr
	conn     *net.UDPConn
	batch    *transport.Batch
	timer    *time.Timer
	duration time.Duration
	fullChan chan struct{}
}

func NewReporter(conf *config.Configuration, tags ...config.Tag) (*Reporter, error) {
	r := new(Reporter)
	err := r.init(conf, tags...)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Reporter) init(conf *config.Configuration, tags ...config.Tag) error {
	addr, err := net.ResolveUDPAddr("udp", conf.Reporter.AgentAddr)
	if err != nil {
		return err
	}

	r.addr = addr
	r.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	ch := make(chan struct{}, 1)
	r.fullChan = ch
	r.duration = conf.Reporter.Duration

	r.batch = transport.NewBatch(conf, ch, tags...)
	r.timer = time.NewTimer(r.duration)

	return nil
}

func (r *Reporter) Start() {
	go r.Run()
}

func (r *Reporter) Run() {
	for {
		select {
		case <-r.timer.C:
			if err := r.Send(); err != nil {
				fmt.Println(err)
				return
			}

			r.timer.Reset(r.duration)
		case <-r.fullChan:
			if !r.timer.Stop() {
				<-r.timer.C
			}

			if err := r.Send(); err != nil {
				fmt.Println(err)
				return
			}

			r.timer.Reset(r.duration)
		}

	}
}

func (r *Reporter) Send() error {
	data, err := r.batch.GetData()
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = r.conn.Write(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (r *Reporter) Store(span span.ToModel) {
	r.batch.Push(span)
}
