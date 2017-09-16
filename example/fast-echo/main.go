package main

import (
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/faceair/fastsocket"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		socket := fastsocket.New(conn, time.Minute)
		err = socket.OnData(func() {
			data, _ := ioutil.ReadAll(conn)
			log.Printf("%v", data)
		}).OnClose(func() {
		}).Listen()
		if err != nil {
			log.Print(err.Error())
		}
	}
}
