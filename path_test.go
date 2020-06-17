package atreugo

import (
	"errors"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

var middlewareFns = []Middleware{
	func(ctx *RequestCtx) error {
		return ctx.ErrorResponse(errors.New("Bad request"), fasthttp.StatusBadRequest)
	},
	func(ctx *RequestCtx) error {
		return ctx.Next()
	},
}

func assertTimeoutFields(t *testing.T, p *Path, timeout time.Duration, msg string, statusCode int) {
	if !p.withTimeout {
		t.Error("Path.withTimeout is not true")
	}

	if p.timeout != timeout {
		t.Errorf("Path.timeout == %d, want %d", p.timeout, timeout)
	}

	if p.timeoutMsg != msg {
		t.Errorf("Path.timeoutMsg == %s, want %s", p.timeoutMsg, msg)
	}

	if p.timeoutCode != statusCode {
		t.Errorf("Path.timeoutCode == %d, want %d", p.timeoutCode, statusCode)
	}
}

func assertHandle(t *testing.T, p *Path) {
	h, _ := p.router.router.Lookup(p.method, p.url, nil)
	if h == nil {
		t.Error("Path not updated")
	}
}

func newTestPath() *Path {
	return &Path{
		router: newRouter(testLog, nil),
		method: fasthttp.MethodGet,
		url:    "/test",
		view:   func(ctx *RequestCtx) error { return nil },
	}
}

func TestPath_Middlewares(t *testing.T) {
	p := newTestPath()
	p.Middlewares(Middlewares{Before: middlewareFns, After: middlewareFns, Skip: middlewareFns})

	if len(p.middlewares.Before) != len(middlewareFns) {
		t.Errorf("Before middlewares are not registered")
	}

	if len(p.middlewares.After) != len(middlewareFns) {
		t.Errorf("After middlewares are not registered")
	}

	if len(p.middlewares.Skip) != len(middlewareFns) {
		t.Errorf("Skip middlewares are not registered")
	}

	assertHandle(t, p)
}

func TestPath_UseBefore(t *testing.T) {
	p := newTestPath()
	p.UseBefore(middlewareFns...)

	if len(p.middlewares.Before) != len(middlewareFns) {
		t.Errorf("Before middlewares are not registered")
	}

	assertHandle(t, p)
}

func TestPath_UseAfter(t *testing.T) {
	p := newTestPath()
	p.UseAfter(middlewareFns...)

	if len(p.middlewares.After) != len(middlewareFns) {
		t.Errorf("After middlewares are not registered")
	}

	assertHandle(t, p)
}

func TestPath_SkipMiddlewares(t *testing.T) {
	p := newTestPath()
	p.SkipMiddlewares(middlewareFns...)

	if len(p.middlewares.Skip) != len(middlewareFns) {
		t.Errorf("Skip middlewares are not registered")
	}

	assertHandle(t, p)
}

func TestPath_Timeout(t *testing.T) {
	p := newTestPath()

	timeout := 10 * time.Millisecond
	msg := "test"

	p.Timeout(timeout, msg)

	assertTimeoutFields(t, p, timeout, msg, fasthttp.StatusRequestTimeout)
	assertHandle(t, p)
}

func TestPath_TimeoutCode(t *testing.T) {
	p := newTestPath()

	timeout := 10 * time.Millisecond
	msg := "test"
	statusCode := 500

	p.TimeoutCode(timeout, msg, statusCode)

	assertTimeoutFields(t, p, timeout, msg, statusCode)
	assertHandle(t, p)
}
