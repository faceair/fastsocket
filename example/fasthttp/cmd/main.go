package main

import (
	"net/http"

	"github.com/faceair/fastsocket/example/fasthttp"
)

func main() {
	fasthttp.ListenAndServe("localhost:8080", func(req *fasthttp.Request, res *fasthttp.Response) {
		res.Status(http.StatusOK)
		res.Write([]byte("hello world"))
	})
}
