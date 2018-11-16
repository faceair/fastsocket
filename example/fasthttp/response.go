package fasthttp

import (
	"fmt"
	"net/http"

	"github.com/faceair/fastsocket"
)

type Response struct {
	Socket *fastsocket.Socket
}

func NewResponse(socket *fastsocket.Socket) *Response {
	return &Response{
		Socket: socket,
	}
}

func (r *Response) Status(code int) {
	status := fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, http.StatusText(code))
	r.Socket.Write([]byte(status))
}

var (
	cColon     = []byte(": ")
	cCRLF      = []byte("\r\n")
	cConnClose = []byte("Connection: close\r\n")
)

func (r *Response) WriteHeader(key, value string) {
	r.Socket.Write([]byte(key))
	r.Socket.Write(cColon)
	r.Socket.Write([]byte(value))
	r.Socket.Write(cCRLF)
}

func (r *Response) Write(body []byte) {
	r.Socket.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))))
	r.Socket.Write(body)
}

func (r *Response) Close() {
	r.Socket.Flush()
	r.Socket.Close()
}
