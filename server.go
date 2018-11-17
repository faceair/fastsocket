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

func NewServer(addrs string) (*Server, error) {
	server := &Server{}
	return server, server.Listen(addrs)
}

type Server struct {
	addr       *Addr
	listenfd   int
	acceptDesc *netpoll.Desc
}

func (s *Server) Listen(addrs string) (err error) {
	s.addr = &Addr{"tcp4", addrs}
	s.listenfd, _, err = tcpCfg.NewFD(s.addr.Network(), s.addr.String())
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
			clientfd, sa, err := syscall.Accept(s.listenfd)
			if err != nil {
				acceptErr <- err
				return
			}
			conn, err := newConn(clientfd, s.addr, sockaddr2Addr(s.addr.Network(), sa))
			if err != nil {
				acceptErr <- err
				return
			}
			workerPool.Schedule(func() { acceptFn(conn) })
		})
		for len(acceptErr) > 0 {
			err = <-acceptErr
		}
		if err != nil {
			if ne, ok := err.(net.Error); ok && (ne.Temporary() || ne.Timeout()) || err == ErrScheduleTimeout {
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
