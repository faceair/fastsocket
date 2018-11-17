package fasthttp

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/faceair/fastsocket"
)

const OptimalBufferSize = 1500

type Handler func(*Request, *Response)

func ListenAndServe(addr string, handler Handler) error {
	fsocket, err := fastsocket.NewServer(addr)
	if err != nil {
		return err
	}
	err = fsocket.Accept(func(conn net.Conn) {
		socket := fastsocket.NewSocket(conn)

		var length, offset int
		var err error

		buffer := make([]byte, OptimalBufferSize)
		req := newRequest(socket)
		res := req.Response

		socket.OnReadable(func() {
			offset, err = socket.Read(buffer[length:])
			if err != nil {
				res.Write([]byte(err.Error()))
				res.Close()
				return
			}
			length += offset
			_, err := req.Parse(buffer[:length])
			if err == ErrMissingData {
				return
			}
			if err != nil {
				res.Write([]byte(err.Error()))
				res.Close()
				return
			}
			handler(req, res)
			res.Close()
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
