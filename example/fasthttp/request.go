package fasthttp

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/faceair/fastsocket"
)

var (
	ErrBadProto    = errors.New("bad protocol")
	ErrMissingData = errors.New("missing data")
	ErrUnsupported = errors.New("unsupported http feature")
)

const (
	eNextHeader int = iota
	eNextHeaderN
	eHeader
	eHeaderValueSpace
	eHeaderValue
	eHeaderValueN
	eMLHeaderStart
	eMLHeaderValue
)

const OptimalBufferSize = 1500

type OnBody func([]byte)

func newRequest(socket *fastsocket.Socket) *Request {
	return &Request{
		Socket:        socket,
		Response:      newResponse(socket),
		Method:        "GET",
		Proto:         "HTTP/1.1",
		Header:        make(http.Header),
		ContentLength: -1,
		buffer:        NewBuffer(make([]byte, OptimalBufferSize)),
		readBody:      false,
	}
}

type Request struct {
	Socket        *fastsocket.Socket
	Response      *Response
	Method        string
	Proto         string
	URL           *url.URL
	Header        http.Header
	ContentLength int64
	Host          string
	RemoteAddr    string
	buffer        *Buffer
	readBody      bool
	onBody        OnBody
}

func (r *Request) onData() (headerEnd, bodyEnd bool, err error) {
	_, err = r.buffer.ReadFrom(r.Socket)
	if err != nil {
		return false, false, err
	}
	if !r.readBody {
		n, err := r.parseHeader(r.buffer.Bytes())
		if err != nil {
			if err == ErrMissingData {
				return false, false, nil
			} else {
				return true, false, err
			}
		}
		r.buffer.Next(n)
		r.readBody = true
	}
	if r.ContentLength >= 0 {
		if int64(r.buffer.Len()) >= r.ContentLength {
			r.onBody(r.buffer.Bytes())
			return true, true, nil
		}
	} else if r.buffer.Len() > 5 {
		if bytes.Equal(r.buffer.Bytes()[r.buffer.Len()-5:], []byte("0\r\n\r\n")) {
			r.onBody(r.buffer.Bytes())
			return true, true, nil
		}
	}
	return true, false, nil
}

func (r *Request) OnBody(onBody OnBody) {
	r.onBody = onBody
}

func (r *Request) parseHeader(input []byte) (int, error) {
	var headers int
	var path int
	var ok bool
	var err error

	total := len(input)

method:
	for i := 0; i < total; i++ {
		switch input[i] {
		case ' ', '\t':
			r.Method = string(input[0:i])
			ok = true
			path = i + 1
			break method
		}
	}

	if !ok {
		return 0, ErrMissingData
	}

	var version int

	ok = false

path:
	for i := path; i < total; i++ {
		switch input[i] {
		case ' ', '\t':
			ok = true
			r.URL, err = url.Parse(string(input[path:i]))
			if err != nil {
				return 0, fmt.Errorf("%s: %s", ErrBadProto, err)
			}
			version = i + 1
			break path
		}
	}

	if !ok {
		return 0, ErrMissingData
	}

	var readN bool

	ok = false
loop:
	for i := version; i < total; i++ {
		c := input[i]

		switch readN {
		case false:
			switch c {
			case '\r':
				r.Proto = string(input[version:i])
				readN = true
			case '\n':
				r.Proto = string(input[version:i])
				headers = i + 1
				ok = true
				break loop
			}
		case true:
			if c != '\n' {
				return 0, fmt.Errorf("%s: %s", ErrBadProto, "missing newline in version")
			}
			headers = i + 1
			ok = true
			break loop
		}
	}

	if !ok {
		return 0, ErrMissingData
	}
	r.Response.proto = r.Proto

	var headerName []byte

	state := eNextHeader
	start := headers

	for i := headers; i < total; i++ {
		switch state {
		case eNextHeader:
			switch input[i] {
			case '\r':
				state = eNextHeaderN
			case '\n':
				return i + 1, nil
			default:
				start = i
				state = eHeader
			}
		case eHeader:
			if input[i] == ':' {
				headerName = input[start:i]
				state = eHeaderValueSpace
			}
		case eHeaderValueSpace:
			switch input[i] {
			case ' ', '\t':
				continue
			}

			start = i
			state = eHeaderValue
		case eHeaderValue:
			switch input[i] {
			case '\r':
				state = eHeaderValueN
			case '\n':
				state = eNextHeader
			default:
				continue
			}

			r.Header.Add(string(headerName), string(input[start:i]))
		case eHeaderValueN:
			if input[i] != '\n' {
				return 0, ErrBadProto
			}
			state = eNextHeader
		case eNextHeaderN:
			if input[i] != '\n' {
				return 0, ErrBadProto
			}

			r.Host = r.Header.Get("Host")
			r.URL.Host = r.Host
			if sLen := r.Header.Get("Content-Length"); len(sLen) > 0 {
				if i, err := strconv.ParseInt(sLen, 10, 0); err == nil {
					r.ContentLength = i
				}
			}
			return i + 1, nil
		}
	}

	return 0, ErrMissingData
}
