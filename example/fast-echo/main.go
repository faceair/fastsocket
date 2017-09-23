package main

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/faceair/fastsocket"
)

func main() {
	server, err := fastsocket.NewServer(":8080")
	if err != nil {
		panic(err)
	}
	exit := make(chan struct{})
	server.Accept(func(conn net.Conn) {
		socket := fastsocket.NewSocket(conn)
		socket.OnReadable(func() {
			b, err := ioutil.ReadAll(socket)
			if err != nil {
				log.Print(err.Error())
			}
			log.Printf("%v", b)
		}).OnClose(func() {}).Listen()
	})
	<-exit
}
