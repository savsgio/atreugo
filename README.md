# Atreugo

[![Build Status](https://travis-ci.org/savsgio/atreugo.svg?branch=master)](https://travis-ci.org/savsgio/atreugo)
[![Coverage Status](https://coveralls.io/repos/github/savsgio/atreugo/badge.svg?branch=master)](https://coveralls.io/github/savsgio/atreugo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/savsgio/atreugo)](https://goreportcard.com/report/github.com/savsgio/atreugo)
[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/savsgio/atreugo/v11)
[![GitHub release](https://img.shields.io/github/release/savsgio/atreugo.svg)](https://github.com/savsgio/atreugo/releases)

High performance and extensible micro web framework with zero memory allocations in hot paths.

It's built on top of [fasthttp](https://github.com/valyala/fasthttp).

## Install

```bash
go get github.com/savsgio/atreugo/v11
```

## Supported Go versions:

- 1.11.x
- 1.12.x
- 1.13.x
- 1.14.x

## Documentation

See: [docs](https://github.com/savsgio/atreugo/tree/master/docs)

## Organization

Find useful libraries like middlewares, websocket, etc.

- See: [Organization](https://github.com/atreugo)

## Feature Overview

- Optimized for speed. Easily handles more than 100K qps and more than 1M concurrent keep-alive connections on modern hardware.

- Optimized for low memory usage.

- Easy 'Connection: Upgrade' support via RequestCtx.Hijack.

- Server provides anti-DoS limits.

- Middlewares support:

  - Before view execution.
  - After view execution.

- Easy routing:

  - Path parameters (mandatories and optionals).
  - Views with timeout.
  - Group paths and middlewares.
  - Static files.
  - Serve one file like pdf, etc.
  - Middlewares for specific views.
  - fasthttp handlers support.
  - net/http handlers support.

- Common responses (also you could use your own responses):
  - JSON
  - HTTP
  - Text
  - Raw
  - File
  - Redirect

## Examples:

Go to [examples](https://github.com/atreugo/examples) to see how to use Atreugo.

## Note:

`*atreugo.RequestCtx` is equal to `*fasthttp.RequestCtx`, but with extra functionalities, so you can use
the same functions of `*fasthttp.RequestCtx`. Don't worry :smile:

## Benchmark

**Best Performance:** Atreugo is **one of the fastest** go web frameworks in the [go-web-framework-benchmark](https://github.com/smallnest/go-web-framework-benchmark).

- Basic Test: The first test case is to mock 0 ms, 10 ms, 100 ms, 500 ms processing time in handlers.

![](https://raw.githubusercontent.com/smallnest/go-web-framework-benchmark/master/benchmark.png)

- Concurrency Test (allocations): In 30 ms processing time, the test result for 100, 1000, 5000 clients is:

\* _Smaller is better_

![](https://raw.githubusercontent.com/smallnest/go-web-framework-benchmark/master/concurrency_alloc.png)

# Contributing

**Feel free to contribute or fork me...** :wink:
