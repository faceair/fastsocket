package main

import (
	"github.com/faceair/fastsocket"
)

func main() {
	server, err := fastsocket.NewServer("0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
	server.Accept(func(clientfd int) {

	})
}
