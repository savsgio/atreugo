package atreugo

import (
	"fmt"
	"reflect"

	"github.com/valyala/fasthttp"
)

func panicf(s string, args ...interface{}) {
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

func isEqual(v1, v2 interface{}) bool {
	return reflect.ValueOf(v1).Pointer() == reflect.ValueOf(v2).Pointer()
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
