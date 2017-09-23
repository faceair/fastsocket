package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/faceair/fastsocket"
)

var HTTPNoContent = []byte("HTTP/1.1 204 No Content\nContent-Length: 0\n\r")

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

	var acceptCount, readableCount, closeCount int
	server.Accept(func(conn net.Conn) {
		acceptCount++
		socket := fastsocket.NewBufferedSocket(conn, 1014, 1024, time.Hour)
		socket.OnReadable(func() {
			readableCount++
			_, err := ioutil.ReadAll(socket)
			if err != nil {
				log.Print(err.Error())
				return
			}
			_, err = socket.Write(HTTPNoContent)
			if err != nil {
				log.Print(err.Error())
				return
			}
			err = socket.Close()
			if err != nil {
				log.Print(err.Error())
				return
			}
			closeCount++
		}).Listen()
	})
	go func() {
		throttle := time.Tick(time.Second)
		for {
			<-throttle
			println(acceptCount, readableCount, closeCount)
		}
	}()
	<-exit
}
