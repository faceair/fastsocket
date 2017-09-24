package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

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
		socket := fastsocket.NewBufferedSocket(conn, 1014, 1024, time.Hour)
		socket.OnReadable(func() {
			io.ReadFull()
			b, err := ioutil.ReadAll(socket)
			if err != nil {
				log.Print(err.Error())
				return
			}
			_, err = socket.Write(b)
			if err != nil {
				log.Print(err.Error())
				return
			}
			err = socket.Close()
			if err != nil {
				log.Print(err.Error())
				return
			}
		}).Listen()
	})
	<-exit
}
