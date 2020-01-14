package atreugo

import (
	"context"
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

// AttachContext attach a context.Context to the RequestCtx
func (ctx *RequestCtx) AttachContext(extraCtx context.Context) {
	ctx.SetUserValue(attachedCtxKey, extraCtx)
}

// AttachedContext returns the attached context.Context if exist
func (ctx *RequestCtx) AttachedContext() context.Context {
	if extraCtx, ok := ctx.RequestCtx.Value(attachedCtxKey).(context.Context); ok {
		return extraCtx
	}

	return nil
}

// Value returns the value associated with this context or extra context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (ctx *RequestCtx) Value(key interface{}) interface{} {
	if val := ctx.RequestCtx.Value(key); val != nil {
		return val
	}

	if extraCtx := ctx.AttachedContext(); extraCtx != nil {
		return extraCtx.Value(key)
	}

	return nil
}
