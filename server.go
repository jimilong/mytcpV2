package main

import (
	"log"
	"net"
	"strings"
	"errors"
	"crypto/md5"
	"crypto/rand"
	"bytes"
	"io"
	"time"
)

var (
	authKey = []byte("abcdef12")
	errServerAuthFailed = errors.New("server auth failed")
)

type Server struct {
	addr          string
	manager       *Manager
	listener      *net.TCPListener
	recvChanSize  int
	sendChanSize  int
	socketBufSize int
}

func NewServer(addr string, rchan int, schan int, sbuf int) *Server {
	return &Server{
		addr:          addr,
		manager:       NewManager(),
		recvChanSize:  rchan,
		sendChanSize:  schan,
		socketBufSize: sbuf,
	}
}

func (s *Server) Serve() {
	var (
		tcpAddr *net.TCPAddr
		conn    *net.TCPConn
		err     error
	)
	if tcpAddr, err = net.ResolveTCPAddr("tcp", s.addr); err != nil {
		log.Printf("net.ResolveTCPAddr(\"tcp\", \"%s\") error(%v) \n", s.addr, err)
		return
	}
	if s.listener, err = net.ListenTCP("tcp", tcpAddr); err != nil {
		log.Printf("net.ListenTCP(\"tcp\", \"%s\") error(%v) \n", tcpAddr, err)
		return
	}
	log.Printf("start tcp listen: \"%s\" \n", s.addr)

	for {
		if conn, err = s.listener.AcceptTCP(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Printf("listener.Accept(\"%s\") error(%v)", s.listener.Addr().String(), err)
			return
		}
		if err = s.serverAuth(conn, authKey); err != nil {
			conn.Close()
			log.Printf("server.serverAuth() error(%v)", err)
			continue
		}
		if err = conn.SetKeepAlive(true); err != nil {
			log.Printf("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(s.socketBufSize); err != nil {
			log.Printf("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(s.socketBufSize); err != nil {
			log.Printf("conn.SetWriteBuffer() error(%v)", err)
			return
		}

		go NewSession(s, conn)
	}
}

func (s *Server) GetSession(sessionID uint64) *Session {
	return s.manager.GetSession(sessionID)
}

func (s *Server) Stop() {
	s.listener.Close()
	s.manager.Close()
}

//鉴权
func (s *Server) serverAuth(conn *net.TCPConn, key []byte) error {
	var buf [md5.Size]byte

	conn.SetDeadline(time.Now().Add(time.Second * 5))
	rand.Read(buf[:8])
	if _, err := conn.Write(buf[:8]); err != nil {
		conn.Close()
		return err
	}

	hash := md5.New()
	hash.Write(buf[:8])
	hash.Write(key)
	verify := hash.Sum(nil)
	if _, err := io.ReadFull(conn, buf[:]); err != nil {
		conn.Close()
		return err
	}
	if !bytes.Equal(verify, buf[:md5.Size]) {
		conn.Close()
		return errServerAuthFailed
	}
	conn.SetDeadline(time.Time{})

	return nil
}