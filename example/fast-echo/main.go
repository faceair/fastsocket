package main

import (
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/faceair/fastsocket"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func main() {
	server, err := fastsocket.NewServer(":8080")
	if err != nil {
		panic(err)
	}
	exit := make(chan struct{})

	server.Accept(func(conn net.Conn) {
		socket := fastsocket.NewSocket(conn)
		socket.OnReadable(func() {
			data := make([]byte, 16)
			_, err := socket.Read(data)
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Printf("%v", data)
			_, err = socket.Write(data)
			if err != nil {
				log.Fatal(err.Error())
			}
		}).Listen()
	})
	<-exit
}
