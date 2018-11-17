package fastsocket

import (
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

func newConn(fd int, lad, rad *Addr) (conn *NonBlockingConn, err error) {
	conn = &NonBlockingConn{
		fd:   fd,
		lad:  lad,
		rad:  rad,
		file: os.NewFile(uintptr(fd), fmt.Sprintf("fsocket.tcp.%d", fd)),
	}
	err = conn.setNoDelay(true)
	if err != nil {
		return
	}
	err = conn.setNonblock(true)
	return
}

type NonBlockingConn struct {
	file *os.File
	fd   int
	lad  *Addr
	rad  *Addr
}

func (c *NonBlockingConn) setNonblock(nonblocking bool) error {
	return syscall.SetNonblock(c.fd, nonblocking)
}

func (c *NonBlockingConn) setNoDelay(noDelay bool) error {
	return syscall.SetsockoptInt(c.fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, boolInt(noDelay))
}

func (c *NonBlockingConn) File() (*os.File, error) {
	return c.file, nil
}

func (c *NonBlockingConn) Read(b []byte) (n int, err error) {
	n, err = fixCount(syscall.Read(c.fd, b))
	if n == 0 && len(b) > 0 && err == nil {
		return 0, io.EOF
	}
	return
}

func (c *NonBlockingConn) Write(b []byte) (n int, err error) {
	n, err = fixCount(syscall.Write(c.fd, b))
	if n != len(b) && err == nil {
		err = io.ErrShortWrite
	}
	return
}

func (c *NonBlockingConn) Close() error {
	return syscall.Close(c.fd)
}

func (c *NonBlockingConn) LocalAddr() net.Addr {
	return c.lad
}

func (c *NonBlockingConn) RemoteAddr() net.Addr {
	return c.rad
}

func (c *NonBlockingConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *NonBlockingConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *NonBlockingConn) SetWriteDeadline(t time.Time) error {
	return nil
}
