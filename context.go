package atreugo

import (
	"sync"

	"github.com/valyala/fasthttp"
)

var requestCtxPool = sync.Pool{
	New: func() interface{} {
		return new(RequestCtx)
	},
}

func (ctx *RequestCtx) reset() {
	ctx.RequestCtx = nil
}

// RequestID return request ID for current context's request
func (ctx *RequestCtx) RequestID() string {
	return ctx.requestID
}

func acquireRequestCtx(ctx *fasthttp.RequestCtx) *RequestCtx {
	actx := requestCtxPool.Get().(*RequestCtx)
	actx.RequestCtx = ctx
	return actx
}

func releaseRequestCtx(actx *RequestCtx) {
	actx.reset()
	requestCtxPool.Put(actx)
}