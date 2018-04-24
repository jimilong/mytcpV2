package main

import (
	"net"
	"bufio"
	"sync"
	"log"
	"errors"
	"sync/atomic"
)

var (
	globalSessionId uint64
	errConnClosing  = errors.New("use of closed network connection")
	errSendBlocking = errors.New("write packet was blocking")
)

type Session struct {
	id        uint64
	manager   *Manager
	conn      *net.TCPConn
	reader    *bufio.Reader
	writer    *bufio.Writer
	recvChan  chan *Packet
	sendChan  chan *Packet
	closeChan chan struct{}
	closeFlag int32
	closeOnce sync.Once
}

func NewSession(server *Server, conn *net.TCPConn) {
	log.Printf("start tcp serve \"%s\" with \"%s\"",
		conn.LocalAddr().String(), conn.RemoteAddr().String())
	session := &Session{
		id:        atomic.AddUint64(&globalSessionId, 1),
		manager:   server.manager,
		conn:      conn,
		reader:    bufio.NewReader(conn),
		writer:    bufio.NewWriter(conn),
		recvChan:  make(chan *Packet, server.recvChanSize),
		sendChan:  make(chan *Packet, server.sendChanSize),
		closeChan: make(chan struct{}),
	}

	go session.readLoop()
	go session.handleLoop()
	go session.writeLoop()

	server.manager.putSession(session)
}

func (s *Session) AsyncSend(p *Packet) error {
	select {
	case s.sendChan <- p:
		return nil
	case <-s.closeChan:
		return errConnClosing
	default:
		return errSendBlocking
	}
}

func (s *Session) readLoop() {
	defer s.Close()

	var (
		packet *Packet
		err    error
	)
	for {
		select {
		case <-s.closeChan:
			return
		default:
		}
		packet = &Packet{} // todo packetPool
		if err = packet.ReadTCP(s.reader); err != nil {
			log.Printf("read tcp error(%v)\n", err)
			break
		}
		s.recvChan <- packet
	}
}

func (s *Session) handleLoop() {
	defer s.Close()

	for {
		select {
		case <-s.closeChan:
			return
		case p, ok := <-s.recvChan:
			if !ok {
				return
			}
			log.Println(p.ToString())
			//todo packet dispatch
		}
	}
}

func (s *Session) writeLoop() {
	defer s.Close()

	var err error
	for {
		select {
		case <-s.closeChan:
			return
		case p, ok := <-s.sendChan:
			if !ok {
				return
			}
			if err = p.WriteTCP(s.writer); err != nil {
				log.Printf("write tcp error (%v) \n", err)
				return
			}
		}
	}
}

func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.closeChan)
		close(s.recvChan)
		close(s.sendChan)
		s.conn.Close()
		s.manager.delSession(s)
		log.Printf("stop tcp serve \"%s\" with \"%s\"",
			s.conn.LocalAddr().String(), s.conn.RemoteAddr().String())
	})
}
