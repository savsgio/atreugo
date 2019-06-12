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
	ctx.RequestCtx = nil
}

// RequestID returns the "X-Request-ID" header value
func (ctx *RequestCtx) RequestID() []byte {
	return ctx.Request.Header.Peek(XRequestIDHeader)
}
