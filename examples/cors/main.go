package main

import (
	"github.com/savsgio/atreugo/v10"
)

func main() {
	config := atreugo.Config{
		Addr: "0.0.0.0:8000",
	}
	server := atreugo.New(config)

	server.UseBefore(corsMiddleware)

	server.POST("/login", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Hello": "World"})
	})

	server.POST("/no-cors", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Hello": "No CORS"})
	})

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
