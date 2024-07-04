package atreugo

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/prefork"
)

func panicf(s string, args ...any) {
	panic(fmt.Sprintf(s, args...))
}

func viewToHandler(view View, errorView ErrorView) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		actx := AcquireRequestCtx(ctx)

		if err := view(actx); err != nil {
			errorView(actx, err, fasthttp.StatusInternalServerError)
		}

		ReleaseRequestCtx(actx)
	}
}

func isEqual(v1, v2 any) bool {
	return reflect.ValueOf(v1).Pointer() == reflect.ValueOf(v2).Pointer()
}

func isNil(v any) bool {
	return reflect.ValueOf(v).IsNil()
}

func middlewaresInclude(ms []Middleware, fn Middleware) bool {
	for _, m := range ms {
		if isEqual(m, fn) {
			return true
		}
	}

	return false
}

func appendMiddlewares(dst, src []Middleware, skip ...Middleware) []Middleware {
	for _, fn := range src {
		if !middlewaresInclude(skip, fn) {
			dst = append(dst, fn)
		}
	}

	return dst
}

func newPreforkServerBase(s *Atreugo) *prefork.Prefork {
	p := &prefork.Prefork{
		Network:          s.cfg.Network,
		Reuseport:        s.cfg.Reuseport,
		RecoverThreshold: runtime.GOMAXPROCS(0) / 2,
		Logger:           s.cfg.Logger,
		ServeFunc:        s.Serve,
	}

	return p
}
