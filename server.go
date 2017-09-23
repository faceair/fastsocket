package fastsocket

import (
	"log"
	"net"
	"syscall"
	"time"

	"github.com/faceair/fastsocket/tcplisten"
	"github.com/mailru/easygo/netpoll"
)

var tcpCfg = &tcplisten.Config{ReusePort: true, DeferAccept: true, FastOpen: true}

func NewServer(addr string) (*Server, error) {
	server := &Server{}
	return server, server.Listen(addr)
}

type Server struct {
	listenfd   int
	acceptDesc *netpoll.Desc
}

func (s *Server) Listen(addr string) (err error) {
	s.listenfd, err = tcpCfg.NewFD("tcp4", addr)
	return
}

func (s *Server) Accept(acceptFn func(net.Conn)) error {
	// Create netpoll descriptor for the listener.
	// We use OneShot here to manually resume events stream when we want to.
	s.acceptDesc = netpoll.NewDesc(uintptr(s.listenfd), netpoll.EventRead|netpoll.EventOneShot)

	// acceptErr is a channel to signal about next incoming connection Accept()
	// results.
	acceptErr := make(chan error, 1)
	poller.Start(s.acceptDesc, func(ev netpoll.Event) {
		err := workerPool.ScheduleTimeout(time.Millisecond, func() {
			clientfd, _, err := syscall.Accept(s.listenfd)
			if err != nil {
				if err != syscall.EAGAIN {
					acceptErr <- err
					return
				}
				acceptErr <- nil
				return
			}
			conn, err := NewConn(clientfd)
			acceptErr <- err
			if err == nil {
				acceptFn(conn)
			}
		})
		if err == nil {
			err = <-acceptErr
		}
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() || ne.Timeout() || err == ErrScheduleTimeout {
				delay := 5 * time.Millisecond
				log.Printf("accept error: %v; retrying in %s", err, delay)
				time.Sleep(delay)
			} else {
				log.Fatalf("accept error: %v", err)
			}
		}

		poller.Resume(s.acceptDesc)
	})
	return nil
}

func (s *Server) Close() error {
	if s.acceptDesc == nil {
		return nil
	}
	poller.Stop(s.acceptDesc)
	s.acceptDesc = nil
	return nil
}
