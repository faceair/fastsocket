package fasthttp

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/faceair/fastsocket"
)

func ListenAndServe(addr string, onRequest OnRequest) error {
	fsocket, err := fastsocket.NewServer(addr)
	if err != nil {
		return err
	}
	err = fsocket.Accept(func(conn net.Conn) {
		socket := fastsocket.NewSocket(conn)
		stream := newStream(socket)
		stream.OnRequest(onRequest)
		socket.OnReadable(func() {
			err := stream.onData()
			if err != nil {
				socket.Close()
			}
		})
		if err = socket.Listen(); err != nil {
			panic(err)
		}
	})
	if err != nil {
		return err
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interrupt

	return fsocket.Close()
}
