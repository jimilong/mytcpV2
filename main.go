package main

import (
	"os"
	"os/signal"
	"syscall"
	"log"
)

func main() {
	serv := NewServer("0.0.0.0:8080", 100, 100, 4096)
	go serv.Serve()

	waitSignal()
	serv.Stop()
}

func waitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s := <-c
		log.Printf("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			return
		default:
			return
		}
	}
}
