package fastsocket

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/mailru/easygo/netpoll"
)

var ErrSocketNeedMoreCallback = errors.New("socket need more callback")
var ErrSocketAlreadyListen = errors.New("socket already listen")

var poller netpoll.Poller

func init() {
	var err error
	poller, err = netpoll.New(nil)
	if err != nil {
		panic(err)
	}
}

func New(conn net.Conn, timeout time.Duration) *Socket {
	return &Socket{
		Conn:    conn,
		timeout: timeout,
	}
}

type Socket struct {
	net.Conn
	io      sync.Mutex
	timeout time.Duration
	desc    *netpoll.Desc
	onData  func()
	onClose func()
}

func (s *Socket) Write(p []byte) (int, error) {
	if err := s.Conn.SetWriteDeadline(CoarseTimeNow().Add(s.timeout)); err != nil {
		return 0, err
	}
	return s.Conn.Write(p)
}

func (s *Socket) Read(p []byte) (int, error) {
	if err := s.Conn.SetReadDeadline(CoarseTimeNow().Add(s.timeout)); err != nil {
		return 0, err
	}
	return s.Conn.Read(p)
}

func (s *Socket) OnData(onData func()) *Socket {
	s.onData = onData
	return s
}

func (s *Socket) OnClose(onClose func()) *Socket {
	s.onClose = onClose
	return s
}

func (s *Socket) Listen() error {
	if s.onData == nil || s.onClose == nil {
		return ErrSocketNeedMoreCallback
	}
	// Create netpoll event descriptor for conn.
	// We want to handle only read events of it.
	if s.desc != nil {
		return ErrSocketAlreadyListen
	}
	s.desc = netpoll.Must(netpoll.HandleRead(s.Conn))

	// Subscribe to events about conn.
	poller.Start(s.desc, func(ev netpoll.Event) {
		if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			// When ReadHup or Hup received, this mean that client has
			// closed at least write end of the connection or connections
			// itself. So we want to stop receive events about such conn
			// and remove it from the chat registry.
			s.onClose()
			return
		}
		// Here we can read some new message from connection.
		// We can not read it right here in callback, because then we will
		// block the poller's inner loop.
		// We do not want to spawn a new goroutine to read single message.
		// But we want to reuse previously spawned goroutine.
		s.onData()
	})

	return nil
}

func (s *Socket) Close() error {
	if s.desc == nil {
		return nil
	}
	poller.Stop(s.desc)
	s.desc = nil
	return s.Conn.Close()
}
