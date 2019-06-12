Atreugo
=======

[![Build Status](https://travis-ci.org/savsgio/atreugo.svg?branch=master)](https://travis-ci.org/savsgio/atreugo)
[![Coverage Status](https://coveralls.io/repos/github/savsgio/atreugo/badge.svg?branch=master)](https://coveralls.io/github/savsgio/atreugo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/savsgio/atreugo)](https://goreportcard.com/report/github.com/savsgio/atreugo)
[![GoDoc](https://godoc.org/github.com/savsgio/atreugo?status.svg)](https://godoc.org/github.com/savsgio/atreugo)
[![GitHub release](https://img.shields.io/github/release/savsgio/atreugo.svg)](https://github.com/savsgio/atreugo/releases)

Micro-framework to make simple the use of routing and middlewares in [fasthttp](https://github.com/valyala/fasthttp).

## Install

```bash
go get github.com/savsgio/atreugo/v8
```

## Benchmark

**Best Performance:** Atreugo is **one of the fastest** go web frameworks in the [go-web-framework-benchmark](https://github.com/smallnest/go-web-framework-benchmark).

- Basic Test: The first test case is to mock 0 ms, 10 ms, 100 ms, 500 ms processing time in handlers.

![](https://raw.githubusercontent.com/smallnest/go-web-framework-benchmark/master/benchmark.png)

- Concurrency Test (allocations): In 30 ms processing time, the tets result for 100, 1000, 5000 clients is:

\* *Smaller is better*

![](https://raw.githubusercontent.com/smallnest/go-web-framework-benchmark/master/concurrency_alloc.png)

## Note:
`*atreugo.RequestCtx` is equal than `*fasthttp.RequestCtx`, but adding extra funtionality, so you can use
the same functions of `*fasthttp.RequestCtx`. Don't worry :smile:

## Example:

```go
package main

import (
	"fmt"
	"time"

	"github.com/savsgio/atreugo/v8"
	"github.com/savsgio/atreugo/v8/middlewares"
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

```

## Useful third-party libraries

- [session](https://github.com/fasthttp/session)
- [websocket](https://github.com/fasthttp/websocket)

Contributing
============

**Feel free to contribute it or fork me...** :wink:
