package atreugo

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func TestRequestCtx_reset(t *testing.T) {
	ctx := new(fasthttp.RequestCtx)
	actx := acquireRequestCtx(ctx)

	actx.reset()

	if actx.RequestCtx != nil {
		t.Errorf("reset() *fasthttp.RequestCtx = %p, want %v", actx.RequestCtx, nil)
	}
}

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
