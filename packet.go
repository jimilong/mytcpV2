package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"mytcpV2/bytes"
	"sync"
)

var (
	errPacketSize = errors.New("packet too large")
)

const (
	maxBodySize = 1024

	packetSize = 2 //包长度
	verSize    = 2
	askSize    = 4
	headerSize = packetSize + verSize + askSize //包头长度

	packetOffset = 0
	verOffset    = 2
	askOffset    = 4
	bodyOffset   = 8
)

const preAllocSize = 64 //todo 预分配buffer大小

var bufpool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBufferSize(preAllocSize)
	},
}

type Packet struct {
	Ver   uint16 `json:"ver"`
	AskID uint32 `json:"ask_id"`
	Body  []byte `json:"body"`
}

func (p *Packet) ReadTCP(rr *bufio.Reader) (err error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer bufpool.Put(buf)
	buf.Reset()
	headerBuf := buf.Peek(headerSize)
	if _, err = io.ReadFull(rr, headerBuf); err != nil {
		return
	}
	allSize := binary.BigEndian.Uint16(headerBuf[packetOffset:verOffset])
	p.Ver = binary.BigEndian.Uint16(headerBuf[verOffset:askOffset])
	p.AskID = binary.BigEndian.Uint32(headerBuf[askOffset:bodyOffset])
	bodySize := allSize - headerSize
	if bodySize > maxBodySize {
		return errPacketSize
	}
	buf.Reset()
	bodyBuf := buf.Peek(int(bodySize))
	if _, err = io.ReadFull(rr, bodyBuf); err != nil {
		return
	}
	p.Body = bodyBuf

	return nil
}

func (p *Packet) WriteTCP(wr *bufio.Writer) (err error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer bufpool.Put(buf)
	buf.Reset()
	allSize := len(p.Body) + headerSize
	packetBuf := buf.Peek(allSize)
	binary.BigEndian.PutUint16(packetBuf[:verOffset], uint16(allSize))
	binary.BigEndian.PutUint16(packetBuf[verOffset:askOffset], p.Ver)
	binary.BigEndian.PutUint32(packetBuf[askOffset:bodyOffset], p.AskID)
	copy(packetBuf[headerSize:], p.Body)
	if _, err = wr.Write(packetBuf); err != nil {
		return
	}
	if err = wr.Flush(); err != nil {
		return
	}

	return nil
}

func (p *Packet) ToString() string {
	return fmt.Sprintf("\n-------- packet --------\nver: %d\naskId: %d\nbody: %s\n-----------------------", p.Ver, p.AskID, p.Body)
}
