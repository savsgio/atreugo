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

func TestPath_Middlewares(t *testing.T) {
	p := new(Path)
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
}

func TestPath_UseBefore(t *testing.T) {
	p := new(Path)
	p.UseBefore(middlewareFns...)

	if len(p.middlewares.Before) != len(middlewareFns) {
		t.Errorf("Before middlewares are not registered")
	}
}

func TestPath_UseAfter(t *testing.T) {
	p := new(Path)
	p.UseAfter(middlewareFns...)

	if len(p.middlewares.After) != len(middlewareFns) {
		t.Errorf("After middlewares are not registered")
	}
}

func TestPath_SkipMiddlewares(t *testing.T) {
	p := new(Path)
	p.SkipMiddlewares(middlewareFns...)

	if len(p.middlewares.Skip) != len(middlewareFns) {
		t.Errorf("Skip middlewares are not registered")
	}
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

func TestPath_Timeout(t *testing.T) {
	p := new(Path)

	timeout := 10 * time.Millisecond
	msg := "test"

	p.Timeout(timeout, msg)

	assertTimeoutFields(t, p, timeout, msg, fasthttp.StatusRequestTimeout)
}

func TestPath_TimeoutCode(t *testing.T) {
	p := new(Path)

	timeout := 10 * time.Millisecond
	msg := "test"
	statusCode := 500

	p.TimeoutCode(timeout, msg, statusCode)

	assertTimeoutFields(t, p, timeout, msg, statusCode)
}
