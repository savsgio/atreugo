package main

import (
	"errors"

	"github.com/savsgio/atreugo/v5"
	"github.com/valyala/fasthttp"
)

func main() {
	config := &atreugo.Config{
		Host: "0.0.0.0",
		Port: 8000,
	}
	server := atreugo.New(config)

	fnMiddlewareOne := func(ctx *atreugo.RequestCtx) (int, error) {
		return fasthttp.StatusOK, nil
	}

	fnMiddlewareTwo := func(ctx *atreugo.RequestCtx) (int, error) {
		// Disable this middleware if you don't want to see this error
		return fasthttp.StatusBadRequest, errors.New("Error example")
	}

	server.UseMiddleware(fnMiddlewareOne, fnMiddlewareTwo)

	server.Path("GET", "/", func(ctx *atreugo.RequestCtx) error {
		return ctx.HTTPResponse("<h1>Atreugo Micro-Framework</h1>")
	})

	server.Path("GET", "/jsonPage", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Atreugo": true})
	})

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
