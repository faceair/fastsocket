package fasthttp

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/faceair/fastsocket"
)

func newResponse(socket *fastsocket.Socket) *Response {
	return &Response{
		Socket: socket,
		proto:  "HTTP/1.1",
		status: http.StatusOK,
		Header: make(http.Header),
	}
}

type Response struct {
	Socket *fastsocket.Socket
	proto  string
	status int
	Header http.Header
	body   []byte
}

func (r *Response) Status(code int) {
	r.status = code
}

func (r *Response) Write(body []byte) {
	r.body = body
}

var (
	cColon = []byte(": ")
	cCRLF  = []byte("\r\n")
)

func (r *Response) Close() {
	r.Socket.Write([]byte(fmt.Sprintf("%s %d %s\r\n", r.proto, r.status, http.StatusText(r.status))))

	if len(r.Header.Get("Content-Type")) == 0 {
		r.Header.Set("Content-Type", "text/html; charset=utf-8")
	}
	r.Header.Set("Content-Length", strconv.Itoa(len(r.body)))

	for key, values := range r.Header {
		for _, value := range values {
			r.Socket.Write([]byte(key))
			r.Socket.Write(cColon)
			r.Socket.Write([]byte(value))
			r.Socket.Write(cCRLF)
		}
	}

	r.Socket.Write(cCRLF)
	r.Socket.Write(r.body)

	r.Socket.Flush()
}
