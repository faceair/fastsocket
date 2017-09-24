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

const DefaultBuffSize = 1024

func NewSocket(conn net.Conn) *Socket {
	return &Socket{
		Conn:    conn,
		Reader:  bufio.NewReaderSize(conn, DefaultBuffSize),
		Writer:  bufio.NewWriterSize(conn, DefaultBuffSize),
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
	Reader              *bufio.Reader
	Writer              *bufio.Writer
	timeout             time.Duration
	readLock, writeLock sync.Mutex
	readDesc, writeDesc *netpoll.Desc
	onReadable, onClose func()
}

func (s *Socket) Read(b []byte) (n int, err error) {
	n, err = s.Reader.Read(b)
	if err == syscall.EAGAIN {
		err = io.EOF
	}
	if err == syscall.EBADF {
		err = io.ErrUnexpectedEOF
		s.Close()
	}
	return
}

func (s *Socket) ReadFull(buf []byte) (n int, err error) {
	min := len(buf)
	for n < min && err == nil {
		var nn int
		nn, err = s.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (s *Socket) Write(b []byte) (n int, err error) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()

	n, err = s.Writer.Write(b)
	if err == syscall.EBADF {
		err = io.ErrUnexpectedEOF
		s.Close()
	}
	return
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
			s.readLock.Lock()
			onReadable()
			s.readLock.Unlock()

			if s.Reader.Buffered() == 0 || s.readDesc == nil {
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
	if s.onReadable == nil {
		panic("socket need OnReadable callback")
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
	if s.readDesc == nil {
		return nil
	}
	err := poller.Stop(s.readDesc)
	if err != nil {
		return err
	}
	s.readDesc = nil
	if s.onClose != nil {
		s.onClose()
	}
	return s.Conn.Close()
}
