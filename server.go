package fastsocket

import (
	"log"
	"net"
	"syscall"
	"time"

	"github.com/mailru/easygo/netpoll"
)

func NewServer(addr string) (*Server, error) {
	server := &Server{}
	return server, server.Listen(addr)
}

type Server struct {
	listenfd   int
	acceptDesc *netpoll.Desc
}

func (s *Server) Listen(addr string) error {
	lfd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return err
	}
	sa4, err := resolveSockAddr4(addr)
	if err != nil {
		return err
	}
	if err = syscall.Bind(lfd, sa4); err != nil {
		syscall.Close(lfd)
		return err
	}
	if err = syscall.Listen(lfd, syscall.SOMAXCONN); err != nil {
		syscall.Close(lfd)
		return err
	}
	if err = syscall.SetNonblock(lfd, true); err != nil {
		syscall.Close(lfd)
		return err
	}

	s.listenfd = lfd
	return nil
}

func (s *Server) Accept(acceptFn func(net.Conn)) error {
	// Create netpoll descriptor for the listener.
	// We use OneShot here to manually resume events stream when we want to.
	s.acceptDesc = netpoll.NewDesc(uintptr(s.listenfd), netpoll.EventRead|netpoll.EventOneShot)

	// acceptErr is a channel to signal about next incoming connection Accept()
	// results.
	acceptErr := make(chan error, 1)
	poller.Start(s.acceptDesc, func(ev netpoll.Event) {
		err := pool.ScheduleTimeout(time.Millisecond, func() {
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
			if isNetTemporary(err) || err != ErrScheduleTimeout {
				delay := 5 * time.Millisecond
				log.Printf("accept error: %v; retrying in %s", err, delay)
				time.Sleep(delay)
			}
			log.Fatalf("accept error: %v", err)
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
