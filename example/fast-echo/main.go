package main

import (
	"bytes"
	"net"

	"github.com/faceair/fastsocket"
)

func main() {
	server, err := fastsocket.NewServer("localhost:8080")
	if err != nil {
		panic(err)
	}
	exit := make(chan struct{})
	server.Accept(func(conn net.Conn) {
		socket := fastsocket.NewSocket(conn)
		data := make([]byte, 32)
		socket.OnReadable(func() {
			n, err := socket.Read(data)
			if err != nil {
				panic(err)
			}
			_, err = socket.Write(data[:n])
			if err != nil {
				panic(err)
			}
			if bytes.ContainsRune(data[:n], '\n') {
				socket.Flush()
			}
		}).Listen()
	})
	<-exit
}
