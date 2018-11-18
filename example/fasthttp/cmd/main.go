package main

import (
	"net/http"

	"github.com/faceair/fastsocket/example/fasthttp"
)

func main() {
	fasthttp.ListenAndServe(":8080", func(req *fasthttp.Request, res *fasthttp.Response) {
		req.OnBody(func(body []byte) {
			res.Status(http.StatusOK)
			res.Write([]byte("hello world"))
			res.Close()
		})
	})
}
