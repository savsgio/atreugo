package main

import (
	"github.com/savsgio/atreugo/v8"
)

func main() {
	config := &atreugo.Config{
		Addr: "0.0.0.0:8000",
	}
	server := atreugo.New(config)

	// Register a route
	server.Path("GET", "/", func(ctx *atreugo.RequestCtx) error {
		return ctx.TextResponse("Hello World")
	})

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
