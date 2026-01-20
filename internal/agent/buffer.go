package agent

import (
	"encoding/json"
	"fmt"
	"net"
	"tracer/pkg/config"
	"tracer/pkg/model"
)

// Buffer receives incoming spans via UDP.
type Buffer struct {
	conn    *net.UDPConn
	addr    *net.UDPAddr
	batchCh chan<- model.Package
}

// NewBuffer creates a new Buffer listening on the specified IP address.
func NewBuffer(ipAddr string, batchCh chan<- model.Package) (*Buffer, error) {
	b := new(Buffer)
	if err := b.init(ipAddr, batchCh); err != nil {
		return nil, err
	}

	return b, nil
}

// init initializes the UDP connection.
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

// close closes the UDP connection.
func (b *Buffer) close() {
	err := b.conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Listen starts listening for incoming UDP packets and processes them.
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

// Enrich adds additional metadata (like agent tags) to the package.
func (b *Buffer) Enrich(batch model.Package) model.Package {
	batch.Process.Tags = append(batch.Process.Tags, config.Tag{
		Key:   "agent.host",
		Value: "node-1",
	})

	return batch
}
