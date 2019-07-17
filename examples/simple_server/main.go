package main

import (
	"fmt"
	"time"

	"github.com/savsgio/atreugo/v8"
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

		return fasthttp.StatusOK, nil
	}

	filters := atreugo.Filters{
		Before: []atreugo.Middleware{
			func(ctx *atreugo.RequestCtx) (int, error) {
				// ... your code

				return fasthttp.StatusOK, nil
			},
		},
		After: []atreugo.Middleware{
			func(ctx *atreugo.RequestCtx) (int, error) {
				// ... your code

				return fasthttp.StatusOK, nil
			},
		},
	}

	server.UseBefore(fnMiddlewareOne)
	server.UseAfter(fnMiddlewareTwo)

	server.PathWithFilters("GET", "/", func(ctx *atreugo.RequestCtx) error {
		return ctx.HTTPResponse("<h1>Atreugo</h1>")
	}, filters)

	server.TimeoutPath("GET", "/jsonPage", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(atreugo.JSON{"Atreugo": true})
	}, 5*time.Second, "Timeout response message")

	v1 := server.NewGroupPath("/v1")
	v1.Path("GET", "/users/:name", func(ctx *atreugo.RequestCtx) error {
		return ctx.TextResponse(fmt.Sprintf("v1 - user name: %s", ctx.UserValue("name")))
	})

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
