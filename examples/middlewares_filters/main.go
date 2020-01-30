package main

import (
	"github.com/savsgio/atreugo/v10"
	"github.com/savsgio/atreugo/v10/middlewares"
)

func main() {
	config := &atreugo.Config{
		Addr: "0.0.0.0:8000",
	}
	server := atreugo.New(config)

	// Register before middlewares
	server.UseBefore(middlewares.RequestIDMiddleware, beforeMiddleware)

	// Register after middlewares
	server.UseAfter(afterMiddleware)

	server.GET("/", func(ctx *atreugo.RequestCtx) error {
		return ctx.TextResponse("Middlewares and view filters")
	}).UseBefore(beforeFilter).UseAfter(afterFilter)

	// Run
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
