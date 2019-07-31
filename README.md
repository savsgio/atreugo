Atreugo
=======

[![Build Status](https://travis-ci.org/savsgio/atreugo.svg?branch=master)](https://travis-ci.org/savsgio/atreugo)
[![Coverage Status](https://coveralls.io/repos/github/savsgio/atreugo/badge.svg?branch=master)](https://coveralls.io/github/savsgio/atreugo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/savsgio/atreugo)](https://goreportcard.com/report/github.com/savsgio/atreugo)
[![GoDoc](https://godoc.org/github.com/savsgio/atreugo?status.svg)](https://godoc.org/github.com/savsgio/atreugo)
[![GitHub release](https://img.shields.io/github/release/savsgio/atreugo.svg)](https://github.com/savsgio/atreugo/releases)

High performance and extensible micro web framework, build on top of [fasthttp](https://github.com/valyala/fasthttp).

## Install

```bash
go get github.com/savsgio/atreugo/v8
```

<!-- ## Documentation

See: [docs]() -->

## Examples:

Go [examples](https://github.com/savsgio/atreugo/tree/master/examples) directory to see how to use Atreugo.


## Note:
`*atreugo.RequestCtx` is equal than `*fasthttp.RequestCtx`, but adding extra funtionality, so you can use
the same functions of `*fasthttp.RequestCtx`. Don't worry :smile:


## Benchmark

**Best Performance:** Atreugo is **one of the fastest** go web frameworks in the [go-web-framework-benchmark](https://github.com/smallnest/go-web-framework-benchmark).

- Basic Test: The first test case is to mock 0 ms, 10 ms, 100 ms, 500 ms processing time in handlers.

![](https://raw.githubusercontent.com/smallnest/go-web-framework-benchmark/master/benchmark.png)

- Concurrency Test (allocations): In 30 ms processing time, the tets result for 100, 1000, 5000 clients is:

\* *Smaller is better*

![](https://raw.githubusercontent.com/smallnest/go-web-framework-benchmark/master/concurrency_alloc.png)


## Useful third-party libraries

- [session](https://github.com/fasthttp/session)
- [websocket](https://github.com/fasthttp/websocket)

Contributing
============

**Feel free to contribute it or fork me...** :wink:
