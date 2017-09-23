package fastsocket

import (
	"bufio"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/mailru/easygo/netpoll"
)

const DefaultReadBuffSize = 1024

func NewSocket(conn net.Conn) *Socket {
	return &Socket{
		Conn:    conn,
		Reader:  bufio.NewReaderSize(conn, DefaultReadBuffSize),
		timeout: time.Minute,
	}
}

func NewBufferedSocket(conn net.Conn, readBuffSize, writeBuffSize int, timeout time.Duration) *Socket {
	return &Socket{
		Conn:    conn,
		Reader:  bufio.NewReaderSize(conn, readBuffSize),
		Writer:  bufio.NewWriterSize(conn, writeBuffSize),
		timeout: timeout,
	}
}

type Socket struct {
	net.Conn
	Reader     *bufio.Reader
	Writer     *bufio.Writer
	timeout    time.Duration
	io         sync.Mutex
	readDesc   *netpoll.Desc
	writeDesc  *netpoll.Desc
	onReadable func()
	onClose    func()
}

func (s *Socket) Read(b []byte) (int, error) {
	return s.Reader.Read(b)
}

func (s *Socket) Write(b []byte) (int, error) {
	return s.Conn.Write(b)
}

func (s *Socket) WriteDelay(b []byte) (int, error) {
	return s.Writer.Write(b)
}

func (s *Socket) Flush() {
	err := s.Writer.Flush()
	if err == io.ErrShortWrite || err == syscall.EAGAIN {
		s.onWritAble(s.Flush)
	}
}

func (s *Socket) onWritAble(onWritAble func()) {
	s.writeDesc = netpoll.Must(netpoll.HandleWriteOnce(s.Conn))
	poller.Start(s.writeDesc, func(ev netpoll.Event) {
		if ev&netpoll.EventHup != 0 {
			s.Close()
			return
		}
		onWritAble()
		poller.Stop(s.writeDesc)
		s.writeDesc = nil
	})
}

func (s *Socket) OnReadable(onReadable func()) *Socket {
	s.onReadable = func() {
		for {
			s.io.Lock()
			onReadable()
			s.io.Unlock()

			if s.Reader.Buffered() == 0 {
				break
			}
		}
	}
	return s
}

func (s *Socket) OnClose(onClose func()) *Socket {
	s.onClose = onClose
	return s
}

func (s *Socket) Listen() error {
	if s.onReadable == nil || s.onClose == nil {
		panic("socket need more callback")
	}
	// Create netpoll event descriptor for conn.
	// We want to handle only read events of it.
	if s.readDesc != nil {
		panic("socket already listen")
	}
	s.readDesc = netpoll.Must(netpoll.HandleRead(s.Conn))

	// Subscribe to events about conn.
	poller.Start(s.readDesc, func(ev netpoll.Event) {
		if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			// When ReadHup or Hup received, this mean that client has
			// closed at least write end of the connection or connections
			// itself. So we want to stop receive events about such conn.
			s.Close()
			return
		}
		// Here we can read some new message from connection.
		// We can not read it right here in callback, because then we will
		// block the poller's inner loop.
		// We do not want to spawn a new goroutine to read single message.
		// But we want to reuse previously spawned goroutine.
		workerPool.Schedule(s.onReadable)
	})

	return nil
}

func (s *Socket) Close() error {
	err := poller.Stop(s.readDesc)
	if err != nil {
		return err
	}
	s.readDesc = nil
	s.onClose()
	return s.Conn.Close()
}
