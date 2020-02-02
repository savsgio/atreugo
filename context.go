package atreugo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
)

var attachedCtxKey = fmt.Sprintf("__attachedCtx::%s__", gotils.RandBytes(make([]byte, 15)))

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

// Next pass control to the next middleware/view function
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
//
// WARNING: The extra context could not be itself
func (ctx *RequestCtx) AttachContext(extraCtx context.Context) {
	if extraCtx == ctx {
		panic("could not attach to itself")
	}

	if previousCtx := ctx.AttachedContext(); previousCtx != nil {
		extraCtx = combinedContext{extraCtx, previousCtx}
	}

	ctx.SetUserValue(attachedCtxKey, extraCtx)
}

// AttachedContext returns the attached context.Context if exist
func (ctx *RequestCtx) AttachedContext() context.Context {
	if extraCtx, ok := ctx.UserValue(attachedCtxKey).(context.Context); ok {
		return extraCtx
	}

	return nil
}

// Value returns the value associated with attached context or this context for key,
// or nil if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
//
// WARNING: The provided key should not be of type string or any other built-in
// to avoid extra allocating when assigning to an interface{}, context keys often
// have concrete type struct{}. Alternatively, exported context key variables' static
// type should be a pointer or interface.
//
// If the key is of type string, try to use:
// 		ctx.SetUserValue("myKey", "myValue")
//		ctx.UserValue("myKey")
//
// instead of:
// 		ctx.AttachContext(context.WithValue(context.Background(), "myKey", "myValue"))
//		ctx.Value("myKey")
//
// to avoid extra allocation
func (ctx *RequestCtx) Value(key interface{}) interface{} {
	if !ctx.searchingOnAttachedCtx {
		if extraCtx := ctx.AttachedContext(); extraCtx != nil {
			ctx.searchingOnAttachedCtx = true
			val := extraCtx.Value(key)
			ctx.searchingOnAttachedCtx = false

			return val
		}
	}

	return ctx.RequestCtx.Value(key)
}

// combinedContext combines the values of two contexts by first looking the Value in top, then in bottom.
// The rest of the methods (Deadline etc) are simply delegated to top
type combinedContext struct {
	top    context.Context
	bottom context.Context
}

func (c combinedContext) Deadline() (deadline time.Time, ok bool) {
	return c.top.Deadline()
}

func (c combinedContext) Done() <-chan struct{} {
	return c.top.Done()
}

func (c combinedContext) Err() error {
	return c.top.Err()
}

// First look the key in top and if nil, look in bottom
func (c combinedContext) Value(key interface{}) interface{} {
	if value := c.top.Value(key); value != nil {
		return value
	}
	return c.bottom.Value(key)
}
