package agent

import (
	"encoding/json"
	"fmt"
	"net"
	"tracer/pkg/config"
	"tracer/pkg/model"
)

// Buffer 接受桶
type Buffer struct {
	conn    *net.UDPConn
	addr    *net.UDPAddr
	batchCh chan<- model.Package
}

func NewBuffer(ipAddr string, batchCh chan<- model.Package) (*Buffer, error) {
	b := new(Buffer)
	if err := b.init(ipAddr, batchCh); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Buffer) init(ipAddr string, batchCh chan<- model.Package) error {
	addr, err := net.ResolveUDPAddr("udp", ipAddr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	b.conn = conn
	b.addr = addr
	b.batchCh = batchCh

	return nil
}

func (b *Buffer) close() {
	err := b.conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (b *Buffer) Listen() error {
	for {
		buf := make([]byte, 65535)
		n, _, err := b.conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		var batch model.Package
		if err := json.Unmarshal(buf[:n], &batch); err != nil {
			return err
		}

		batch = b.Enrich(batch)

		b.batchCh <- batch
	}
}

func (b *Buffer) Enrich(batch model.Package) model.Package {
	batch.Process.Tags = append(batch.Process.Tags, config.Tag{
		Key:   "agent.host",
		Value: "node-1",
	})

	return batch
}
