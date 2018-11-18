package fasthttp

import (
	"github.com/faceair/fastsocket"
)

type OnRequest func(*Request, *Response)

func newStream(socket *fastsocket.Socket) *Stream {
	return &Stream{socket: socket}
}

type Stream struct {
	socket      *fastsocket.Socket
	request     *Request
	onRequest   OnRequest
	callRequest bool
}

func (s *Stream) onData() error {
	if s.request == nil {
		s.request = newRequest(s.socket)
		s.callRequest = true
	}
	headerEnd, bodyEnd, err := s.request.onData()
	if err != nil {
		return err
	}
	if headerEnd && s.callRequest {
		s.onRequest(s.request, s.request.Response)
		s.callRequest = false
	}
	if bodyEnd {
		s.request = nil
	}
	return nil
}

func (s *Stream) OnRequest(onRequest OnRequest) {
	s.onRequest = onRequest
}
