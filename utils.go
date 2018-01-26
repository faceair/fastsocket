package fastsocket

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

func sockaddr2Addr(network string, sa syscall.Sockaddr) Addr {
	switch tcpAddr := sa.(type) {
	case *syscall.SockaddrInet4:
		return Addr{network, fmt.Sprintf("%s:%d", tcpAddr.Addr, tcpAddr.Port)}
	case *syscall.SockaddrInet6:
		return Addr{network, fmt.Sprintf("%s:%d", tcpAddr.Addr, tcpAddr.Port)}
	default:
		return Addr{network, ""}
	}
}

func getAddr(network, addr string) (Addr, error) {
	if network != "tcp4" && network != "tcp6" {
		return Addr{network, addr}, errors.New("only tcp4 and tcp6 network is supported")
	}

	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return Addr{network, addr}, err
	}

	switch network {
	case "tcp4":
		return Addr{network, fmt.Sprintf("%s:%d", tcpAddr.IP.To4(), tcpAddr.Port)}, nil
	case "tcp6":
		return Addr{network, fmt.Sprintf("%s:%d", tcpAddr.IP.To16(), tcpAddr.Port)}, nil
	default:
		return Addr{network, addr}, nil
	}
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
