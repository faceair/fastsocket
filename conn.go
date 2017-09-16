package fastsocket

import (
	"net"
	"os"
	"time"
)

func NewConn(fd int) *Conn {
	return &Conn{
		fd: fd,
		f:  os.NewFile(uintptr(fd), "netlink"),
	}
}

type Conn struct {
	f  *os.File
	fd int
}

func (c *Conn) File() (*os.File, error) {
	return c.f, nil
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return c.f.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.f.Write(b)
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
