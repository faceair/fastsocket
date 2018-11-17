package fastsocket

import (
	"fmt"
	"net"
	"syscall"
)

func sockaddr2Addr(network string, sa syscall.Sockaddr) *Addr {
	tcpAddr := sa.(*syscall.SockaddrInet4)
	ipv4 := net.IPv4(tcpAddr.Addr[0], tcpAddr.Addr[1], tcpAddr.Addr[2], tcpAddr.Addr[3])
	return &Addr{network, fmt.Sprintf("%s:%d", ipv4, tcpAddr.Port)}
}

type Addr struct {
	net string
	ad  string
}

func (a Addr) Network() string {
	return a.net
}
func (a Addr) String() string {
	return a.ad
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Many functions in package syscall return a count of -1 instead of 0.
// Using fixCount(call()) instead of call() corrects the count.
func fixCount(n int, err error) (int, error) {
	if n < 0 {
		n = 0
	}
	return n, err
}
