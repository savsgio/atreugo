package main

import (
	"errors"

	"github.com/erikdubbelboer/fasthttp"
	"github.com/savsgio/atreugo"
)

func main() {
	server := atreugo.New()

	fnMiddlewareOne := func(ctx *fasthttp.RequestCtx) (int, error) {
		return fasthttp.StatusOK, nil
	}

	fnMiddlewareTwo := func(ctx *fasthttp.RequestCtx) (int, error) {
		println("This middleware launch an error...")
		return fasthttp.StatusBadRequest, errors.New("Fake error")
	}

	server.UseMiddleware(fnMiddlewareOne, fnMiddlewareTwo)

	server.Path("GET", "/", func(ctx *fasthttp.RequestCtx) error {
		return atreugo.HttpResponse(ctx, []byte("<h1>Atreugo Micro-Framework</h1>"))
	})

	server.Path("GET", "/jsonPage", func(ctx *fasthttp.RequestCtx) error {
		return atreugo.JsonResponse(ctx, atreugo.Json{"Atreugo": true})
	})

	server.ListenAndServe("0.0.0.0", 8000)
}
