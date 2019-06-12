package main

import (
	"fmt"
	"time"

	"github.com/savsgio/atreugo/v7"
	"github.com/savsgio/atreugo/v7/middlewares"
	"github.com/valyala/fasthttp"
)

func main() {
	config := &atreugo.Config{
		Host: "0.0.0.0",
		Port: 8000,
	}
	server := atreugo.New(config)

	fnMiddlewareOne := func(ctx *atreugo.RequestCtx) (int, error) {
		// ... your code
		return fasthttp.StatusOK, nil
	}

	fnMiddlewareTwo := func(ctx *atreugo.RequestCtx) (int, error) {
		// ... your code

		// Disable this middleware if you don't want to see this error
		return fasthttp.StatusBadRequest, fmt.Errorf("%s - Error example", ctx.RequestID())
	}

	server.UseBefore(middlewares.RequestIDMiddleware, fnMiddlewareOne)
	server.UseAfter(fnMiddlewareTwo)

	server.Path("GET", "/", func(ctx *atreugo.RequestCtx) error {
		return ctx.HTTPResponse("<h1>Atreugo</h1>")
	})

	server.TimeoutPath("GET", "/jsonPage", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Atreugo": true})
	}, 5*time.Second, "Timeout response message")

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
