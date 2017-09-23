package fastsocket

import (
	"net"
	"syscall"
)

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

func resolveSockAddr4(netaddr string) (syscall.Sockaddr, error) {
	addr, err := net.ResolveTCPAddr("tcp4", netaddr)
	if err != nil {
		return nil, err
	}
	ip := addr.IP
	if len(ip) == 0 {
		ip = net.IPv4zero
	}
	sa4 := &syscall.SockaddrInet4{Port: addr.Port}
	copy(sa4.Addr[:], ip.To4())
	return sa4, nil
}
