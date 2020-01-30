package atreugo

import (
	"reflect"

	"github.com/valyala/fasthttp"
)

// execute executes all middlewares + view with the given request context
func execute(ctx *RequestCtx, hs []Middleware) error {
	for _, h := range hs {
		if err := h(ctx); err != nil {
			return err
		}

		if !ctx.next {
			return nil
		}

		ctx.next = false
	}

	return nil
}

func viewToHandler(view View, errorView ErrorView) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		actx := acquireRequestCtx(ctx)

		if err := view(actx); err != nil {
			errorView(actx, err, fasthttp.StatusInternalServerError)
		}

		releaseRequestCtx(actx)
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
