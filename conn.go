package fastsocket

import (
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

func NewConn(fd int) (*Conn, error) {
	conn := &Conn{
		fd: fd,
		f:  os.NewFile(uintptr(fd), "tcp"),
	}
	err := conn.SetNoDelay(true)
	if err != nil {
		return conn, err
	}
	return conn, conn.setNonblock(true)
}

// Conn is Non-blocking io connection
type Conn struct {
	f  *os.File
	fd int
}

func (c *Conn) setNonblock(nonblocking bool) error {
	return syscall.SetNonblock(c.fd, nonblocking)
}

func (c *Conn) SetNoDelay(noDelay bool) error {
	return syscall.SetsockoptInt(c.fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, boolInt(noDelay))
}

func (c *Conn) File() (*os.File, error) {
	return c.f, nil
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = fixCount(syscall.Read(c.fd, b))
	if n == 0 && len(b) > 0 && err == nil {
		return 0, io.EOF
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return fixCount(syscall.Write(c.fd, b))
}

func (c *Conn) Close() error {
	return c.f.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return nil
}

func (c *Conn) RemoteAddr() net.Addr {
	return nil
}

func (c *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}
