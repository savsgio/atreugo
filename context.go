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

func acquireRequestCtx(ctx *fasthttp.RequestCtx) *RequestCtx {
	actx := requestCtxPool.Get().(*RequestCtx)
	actx.RequestCtx = ctx
	return actx
}

func releaseRequestCtx(actx *RequestCtx) {
	actx.reset()
	requestCtxPool.Put(actx)
}

func (ctx *RequestCtx) reset() {
	ctx.next = false
	ctx.skipView = false
	ctx.RequestCtx = nil
}

// RequestID returns the "X-Request-ID" header value
func (ctx *RequestCtx) RequestID() []byte {
	return ctx.Request.Header.Peek(XRequestIDHeader)
}

// Next pass control to the next middleware/filter/view function
func (ctx *RequestCtx) Next() error {
	ctx.next = true
	return nil
}

// SkipView sets flag to skip view execution in the current request
//
// Use it in before middlewares
func (ctx *RequestCtx) SkipView() {
	ctx.skipView = true
}
