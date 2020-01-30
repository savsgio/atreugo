package main

import (
	"github.com/savsgio/atreugo/v10"
)

func main() {
	config := &atreugo.Config{
		Addr: "0.0.0.0:8000",
	}
	server := atreugo.New(config)

	server.GET("/", func(ctx *atreugo.RequestCtx) error {
		return ctx.TextResponse("Hello World")
	})

	v1 := server.NewGroupPath("/v1")
	v1.GET("/", func(ctx *atreugo.RequestCtx) error {
		return ctx.TextResponse("Hello V1 Group")
	})

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
