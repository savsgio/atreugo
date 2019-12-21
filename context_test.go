package atreugo

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func Test_acquireRequestCtx(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := acquireRequestCtx(ctx)

	if actx.RequestCtx != ctx {
		t.Errorf("acquireRequestCtx() = %p, want %p", actx.RequestCtx, ctx)
	}
}

func Test_releaseRequestCtx(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := acquireRequestCtx(ctx)

	releaseRequestCtx(actx)

	if actx.RequestCtx != nil {
		t.Errorf("releaseRequestCtx() *fasthttp.RequestCtx = %p, want %v", actx.RequestCtx, nil)
	}
}

func TestRequestCtx_reset(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := acquireRequestCtx(ctx)
	actx.Next()
	actx.SkipView()

	actx.reset()

	if actx.next {
		t.Errorf("reset() next is not 'false'")
	}

	if actx.skipView {
		t.Errorf("reset() skipView is not 'false'")
	}

	if actx.RequestCtx != nil {
		t.Errorf("reset() *fasthttp.RequestCtx = %p, want %v", actx.RequestCtx, nil)
	}
}

func TestRequestCtx_RequestID(t *testing.T) {
	value := "123bnj3r2j3rj23"

	ctx := new(RequestCtx)
	ctx.RequestCtx = new(fasthttp.RequestCtx)
	ctx.Request.Header.Set(XRequestIDHeader, value)

	currentValue := string(ctx.RequestID())
	if currentValue != value {
		t.Errorf("ctx.RequestID() = '%s', want '%s'", currentValue, value)
	}
}

func Test_Next(t *testing.T) {
	ctx := acquireRequestCtx(new(fasthttp.RequestCtx))

	if err := ctx.Next(); err != nil {
		t.Errorf("ctx.Next() unexpected error: %v", err)
	}
}

func Test_SkipView(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := acquireRequestCtx(ctx)

	actx.SkipView()

	if !actx.skipView {
		t.Error("ctx.SkipView() is not true")
	}
}
