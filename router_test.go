package atreugo

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	fastrouter "github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var httpMethods = []string{
	fasthttp.MethodGet,
	fasthttp.MethodHead,
	fasthttp.MethodPost,
	fasthttp.MethodPut,
	fasthttp.MethodPatch,
	fasthttp.MethodDelete,
	fasthttp.MethodConnect,
	fasthttp.MethodOptions,
	fasthttp.MethodTrace,
	fastrouter.MethodWild,
}

func testRouter() *Router {
	return newRouter(testConfig)
}

func randomHTTPMethod() string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(httpMethods)-1)))
	if err != nil {
		panic(err)
	}

	return httpMethods[n.Int64()]
}

func catchPanic(testFunc func()) (recv interface{}) {
	defer func() {
		recv = recover()
	}()

	testFunc()

	return nil
}

func TestRouter_defaultErrorView(t *testing.T) {
	err := errors.New("error")
	statusCode := 500
	ctx := AcquireRequestCtx(new(fasthttp.RequestCtx))

	defaultErrorView(ctx, err, statusCode)

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("Status code == %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusInternalServerError)
	}

	if string(ctx.Response.Body()) != err.Error() {
		t.Errorf("Response body == %s, want %s", ctx.Response.Body(), err.Error())
	}
}

func TestRouter_emptyView(t *testing.T) {
	ctx := AcquireRequestCtx(new(fasthttp.RequestCtx))

	if err := emptyView(ctx); err != nil {
		t.Errorf("emptyView() must returns a nil error")
	}
}

func TestRouter_buildOptionsView(t *testing.T) {
	url := "/"
	paths := map[string][]string{
		fasthttp.MethodGet:  {url},
		fasthttp.MethodPost: {url},
	}

	var errOptionsView error

	headerKey := "key"
	headerValue := "value"
	optionsView := func(ctx *RequestCtx) error {
		ctx.Response.Header.Set(headerKey, headerValue)

		return errOptionsView
	}

	h := buildOptionsView(url, optionsView, paths)

	ctx := AcquireRequestCtx(new(fasthttp.RequestCtx))

	if err := h(ctx); !errors.Is(err, errOptionsView) {
		t.Errorf("Error == %v, want %v", err, errOptionsView)
	}

	wantAllowHeader := "GET, POST"
	allowHeaderValue := string(ctx.Response.Header.Peek("Allow"))

	if allowHeaderValue != wantAllowHeader {
		t.Errorf("Allow header == %s, want %s", allowHeaderValue, wantAllowHeader)
	}

	customHeaderValue := string(ctx.Response.Header.Peek(headerKey))
	if customHeaderValue != headerValue {
		t.Errorf("Header '%s' == %s, want %s", headerKey, customHeaderValue, headerValue)
	}

	h = buildOptionsView(url, optionsView, nil)

	if err := h(ctx); err != nil {
		t.Errorf("emptyView() must returns a nil error")
	}

	wantAllowHeader = "OPTIONS"
	allowHeaderValue = string(ctx.Response.Header.Peek("Allow"))

	if allowHeaderValue != wantAllowHeader {
		t.Errorf("Allow header == %s, want %s", allowHeaderValue, wantAllowHeader)
	}
}

func TestRouter_newRouter(t *testing.T) {
	r := newRouter(testConfig)

	if r.router == nil {
		t.Error("Router instance is nil")
	}

	if r.routerMutable {
		t.Error("Router routerMutable is true")
	}

	if !isEqual(r.errorView, testConfig.ErrorView) {
		t.Errorf("Router errorView == %p, want %p", r.errorView, testConfig.ErrorView)
	}

	if !r.handleOPTIONS {
		t.Error("Router handleOPTIONS is false")
	}
}

func TestRouter_mutable(t *testing.T) {
	handler := func(ctx *fasthttp.RequestCtx) {}

	r := testRouter()
	r.router.GET("/", handler)

	values := []bool{true, false, false, true, true, false}

	for _, v := range values {
		r.mutable(v)

		if r.routerMutable != v {
			t.Errorf("Router.routerMutable == %v, want %v", r.routerMutable, v)
		}

		err := catchPanic(func() {
			r.router.GET("/", handler)
		})

		isMutable := err == nil
		if isMutable != v {
			t.Errorf("Router internal mutable == %v, want %v", isMutable, v)
		}
	}
}

func TestRouter_buildMiddlewares(t *testing.T) {
	middleware1 := func(ctx *RequestCtx) error { return ctx.Next() }
	middleware2 := func(ctx *RequestCtx) error { return ctx.Next() }
	middleware3 := func(ctx *RequestCtx) error { return ctx.Next() }
	middleware4 := func(ctx *RequestCtx) {}

	middle := Middlewares{
		Before: []Middleware{middleware1, middleware2},
		After:  []Middleware{middleware3},
		Final:  []FinalMiddleware{middleware4},
	}
	m := Middlewares{
		Skip: []Middleware{middleware1},
	}

	for _, debug := range []bool{false, true} {
		s := New(Config{
			Debug: debug,
		})
		s.Middlewares(middle)

		result := s.buildMiddlewares(m)

		wantSkipLen := len(m.Skip) + len(middle.Skip)
		if len(result.Skip) != wantSkipLen {
			t.Errorf("Middlewares.Skip length == %d, want %d", len(result.Skip), wantSkipLen)
		}

		if wantBeforeLen := len(middle.Before) - len(m.Skip); len(result.Before) != wantBeforeLen {
			t.Errorf("Middlewares.Before length == %d, want %d", len(result.Before), wantBeforeLen)
		}

		wantAfterLen := len(middle.After)
		if len(result.After) != wantAfterLen {
			t.Errorf("Middlewares.After length == %d, want %d", len(result.After), wantAfterLen)
		}

		if wantFinalLen := len(middle.Final); len(result.Final) != wantFinalLen {
			t.Errorf("Middlewares.Final length == %d, want %d", len(result.Final), wantFinalLen)
		}
	}
}

func TestRouter_handlerExecutionChain(t *testing.T) { //nolint:funlen
	s := New(testConfig)

	method := randomHTTPMethod()
	url := "/foo"

	skipMiddlewareGlobalCalled := false
	skipMiddlewareGroupCalled := false
	viewCalled := false

	callOrder := map[string]int{
		"globalBefore": 0,
		"groupBefore":  0,
		"viewBefore":   0,
		"viewAfter":    0,
		"groupAfter":   0,
		"globalAfter":  0,
		"viewFinal":    0,
		"groupFinal":   0,
		"globalFinal":  0,
	}

	wantOrder := map[string]int{
		"globalBefore": 1,
		"groupBefore":  2,
		"viewBefore":   3,
		"viewAfter":    4,
		"groupAfter":   5,
		"globalAfter":  6,
		"viewFinal":    7,
		"groupFinal":   8,
		"globalFinal":  9,
	}

	index := 0

	skipMiddlewareGlobal := func(ctx *RequestCtx) error {
		skipMiddlewareGlobalCalled = true

		return ctx.Next()
	}
	skipMiddlewareGroup := func(ctx *RequestCtx) error {
		skipMiddlewareGroupCalled = true

		return ctx.Next()
	}

	s.UseBefore(func(ctx *RequestCtx) error {
		index++
		callOrder["globalBefore"] = index

		return ctx.Next()
	}, skipMiddlewareGlobal)
	s.UseAfter(func(ctx *RequestCtx) error {
		index++
		callOrder["globalAfter"] = index

		return ctx.Next()
	})
	s.UseFinal(func(ctx *RequestCtx) {
		index++
		callOrder["globalFinal"] = index
	})

	v1 := s.NewGroupPath("/v1")
	v1.UseBefore(func(ctx *RequestCtx) error {
		index++
		callOrder["groupBefore"] = index

		return ctx.Next()
	})
	v1.UseAfter(func(ctx *RequestCtx) error {
		index++
		callOrder["groupAfter"] = index

		return ctx.Next()
	}, skipMiddlewareGroup)
	v1.UseFinal(func(ctx *RequestCtx) {
		index++
		callOrder["groupFinal"] = index
	})

	v1.SkipMiddlewares(skipMiddlewareGlobal)

	v1.Path(method, url, func(ctx *RequestCtx) error {
		viewCalled = true

		return nil
	}).UseBefore(func(ctx *RequestCtx) error {
		index++
		callOrder["viewBefore"] = index

		return ctx.Next()
	}).UseAfter(func(ctx *RequestCtx) error {
		index++
		callOrder["viewAfter"] = index

		return ctx.Next()
	}).UseFinal(func(ctx *RequestCtx) {
		index++
		callOrder["viewFinal"] = index
	}).SkipMiddlewares(skipMiddlewareGroup)

	ctx := new(fasthttp.RequestCtx)
	h, _ := s.router.Lookup(method, "/v1"+url, ctx)

	if h == nil {
		t.Fatal("Registered handler is nil")
	}

	h(ctx)

	for k, v := range wantOrder {
		if callOrder[k] != v {
			t.Errorf("%s executed at %d, want %d", k, callOrder[k], v)
		}
	}

	if !viewCalled {
		t.Error("View is not called")
	}

	if skipMiddlewareGlobalCalled {
		t.Error("Skip middleware (global) has been called")
	}

	if skipMiddlewareGroupCalled {
		t.Error("Skip middleware (group) has been called")
	}
}

func TestRouter_getGroupFullPath(t *testing.T) {
	r := testRouter()
	foo := r.NewGroupPath("/foo")
	bar := foo.NewGroupPath("/bar")
	buz := bar.NewGroupPath("/buz")

	path := "/atreugo/"

	fullPath := buz.getGroupFullPath(path)
	expected := "/foo/bar/buz/atreugo/"

	if fullPath != expected {
		t.Errorf("Router.getGroupFullPath == %s, want %s", fullPath, expected)
	}

	fullPath = bar.getGroupFullPath(path)
	expected = "/foo/bar/atreugo/"

	if fullPath != expected {
		t.Errorf("Router.getGroupFullPath == %s, want %s", fullPath, expected)
	}

	fullPath = foo.getGroupFullPath(path)
	expected = "/foo/atreugo/"

	if fullPath != expected {
		t.Errorf("Router.getGroupFullPath == %s, want %s", fullPath, expected)
	}
}

func TestRouter_handler(t *testing.T) { //nolint:funlen,maintidx
	type counter struct {
		viewCalled            bool
		beforeMiddlewares     int
		beforeViewMiddlewares int
		afterViewMiddlewares  int
		afterMiddlewares      int
		finalViewMiddlewares  int
		finalMiddlewares      int
	}

	type args struct {
		viewFn      View
		before      []Middleware
		after       []Middleware
		final       []FinalMiddleware
		middlewares Middlewares
	}

	type want struct {
		statusCode int
		counter    counter
	}

	handlerCounter := counter{}
	err := errors.New("test error")

	viewFn := func(ctx *RequestCtx) error {
		handlerCounter.viewCalled = true

		return ctx.TextResponse("Ok")
	}
	before := []Middleware{
		func(ctx *RequestCtx) error {
			handlerCounter.beforeMiddlewares++

			return ctx.Next()
		},
	}
	after := []Middleware{
		func(ctx *RequestCtx) error {
			handlerCounter.afterMiddlewares++

			return ctx.Next()
		},
	}
	final := []FinalMiddleware{
		func(ctx *RequestCtx) {
			handlerCounter.finalMiddlewares++
		},
	}

	middlewares := Middlewares{
		Before: []Middleware{
			func(ctx *RequestCtx) error {
				handlerCounter.beforeViewMiddlewares++

				return ctx.Next()
			},
		},
		After: []Middleware{
			func(ctx *RequestCtx) error {
				handlerCounter.afterViewMiddlewares++

				return ctx.Next()
			},
		},
		Final: []FinalMiddleware{
			func(ctx *RequestCtx) {
				handlerCounter.finalViewMiddlewares++
			},
		},
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "AllOk",
			args: args{
				viewFn:      viewFn,
				before:      before,
				after:       after,
				final:       final,
				middlewares: middlewares,
			},
			want: want{
				statusCode: fasthttp.StatusOK,
				counter: counter{
					viewCalled:            true,
					beforeMiddlewares:     len(before),
					beforeViewMiddlewares: len(middlewares.Before),
					afterViewMiddlewares:  len(middlewares.After),
					afterMiddlewares:      len(after),
					finalViewMiddlewares:  len(middlewares.Final),
					finalMiddlewares:      len(final),
				},
			},
		},
		{
			name: "SkipView",
			args: args{
				viewFn: viewFn,
				before: []Middleware{
					func(ctx *RequestCtx) error {
						handlerCounter.beforeMiddlewares++
						ctx.SkipView()

						return ctx.Next()
					},
				},
				after:       after,
				final:       final,
				middlewares: middlewares,
			},
			want: want{
				statusCode: fasthttp.StatusOK,
				counter: counter{
					viewCalled:            false,
					beforeMiddlewares:     len(before),
					beforeViewMiddlewares: len(middlewares.Before),
					afterViewMiddlewares:  len(middlewares.After),
					afterMiddlewares:      len(after),
					finalViewMiddlewares:  len(middlewares.Final),
					finalMiddlewares:      len(final),
				},
			},
		},
		{
			name: "ViewError",
			args: args{
				viewFn: func(ctx *RequestCtx) error {
					return err
				},
				before:      before,
				after:       after,
				final:       final,
				middlewares: middlewares,
			},
			want: want{
				statusCode: fasthttp.StatusInternalServerError,
				counter: counter{
					viewCalled:            false,
					beforeMiddlewares:     len(before),
					beforeViewMiddlewares: len(middlewares.Before),
					afterViewMiddlewares:  0,
					afterMiddlewares:      0,
					finalViewMiddlewares:  len(middlewares.Final),
					finalMiddlewares:      len(final),
				},
			},
		},
		{
			name: "BeforeMiddlewaresError",
			args: args{
				viewFn: viewFn,
				before: []Middleware{
					func(ctx *RequestCtx) error {
						handlerCounter.beforeMiddlewares++

						return ctx.ErrorResponse(err, fasthttp.StatusBadRequest)
					},
				},
				after:       after,
				final:       final,
				middlewares: middlewares,
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
				counter: counter{
					viewCalled:            false,
					beforeMiddlewares:     1,
					beforeViewMiddlewares: 0,
					afterViewMiddlewares:  0,
					afterMiddlewares:      0,
					finalViewMiddlewares:  len(middlewares.Final),
					finalMiddlewares:      len(final),
				},
			},
		},
		{
			name: "BeforeViewError",
			args: args{
				viewFn: viewFn,
				before: before,
				after:  after,
				final:  final,
				middlewares: Middlewares{
					Before: []Middleware{
						func(ctx *RequestCtx) error {
							handlerCounter.beforeViewMiddlewares++

							return ctx.ErrorResponse(err, fasthttp.StatusBadRequest)
						},
					},
					After: []Middleware{
						func(ctx *RequestCtx) error {
							handlerCounter.afterViewMiddlewares++

							return ctx.Next()
						},
					},
					Final: []FinalMiddleware{
						func(ctx *RequestCtx) {
							handlerCounter.finalViewMiddlewares++
						},
					},
				},
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
				counter: counter{
					viewCalled:            false,
					beforeMiddlewares:     len(before),
					beforeViewMiddlewares: 1,
					afterViewMiddlewares:  0,
					afterMiddlewares:      0,
					finalViewMiddlewares:  len(middlewares.Final),
					finalMiddlewares:      len(final),
				},
			},
		},
		{
			name: "AfterViewError",
			args: args{
				viewFn: viewFn,
				before: before,
				after:  after,
				final:  final,
				middlewares: Middlewares{
					Before: []Middleware{
						func(ctx *RequestCtx) error {
							handlerCounter.beforeViewMiddlewares++

							return ctx.Next()
						},
					},
					After: []Middleware{
						func(ctx *RequestCtx) error {
							handlerCounter.afterViewMiddlewares++

							return ctx.ErrorResponse(err, fasthttp.StatusBadRequest)
						},
					},
					Final: []FinalMiddleware{
						func(ctx *RequestCtx) {
							handlerCounter.finalViewMiddlewares++
						},
					},
				},
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
				counter: counter{
					viewCalled:            true,
					beforeMiddlewares:     len(before),
					beforeViewMiddlewares: 1,
					afterViewMiddlewares:  1,
					afterMiddlewares:      0,
					finalViewMiddlewares:  len(middlewares.Final),
					finalMiddlewares:      len(final),
				},
			},
		},
		{
			name: "AfterMiddlewaresError",
			args: args{
				viewFn: viewFn,
				before: before,
				after: []Middleware{
					func(ctx *RequestCtx) error {
						handlerCounter.afterMiddlewares++

						return ctx.ErrorResponse(err, fasthttp.StatusBadRequest)
					},
				},
				final:       final,
				middlewares: middlewares,
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
				counter: counter{
					viewCalled:            true,
					beforeMiddlewares:     len(before),
					beforeViewMiddlewares: len(middlewares.Before),
					afterViewMiddlewares:  len(middlewares.After),
					afterMiddlewares:      1,
					finalViewMiddlewares:  len(middlewares.Final),
					finalMiddlewares:      len(final),
				},
			},
		},
		{
			name: "NoNext",
			args: args{
				viewFn: viewFn,
				before: []Middleware{
					func(ctx *RequestCtx) error {
						handlerCounter.beforeMiddlewares++

						return nil
					},
				},
				after:       after,
				final:       final,
				middlewares: middlewares,
			},
			want: want{
				statusCode: fasthttp.StatusOK,
				counter: counter{
					viewCalled:            false,
					beforeMiddlewares:     1,
					beforeViewMiddlewares: 0,
					afterViewMiddlewares:  0,
					afterMiddlewares:      0,
					finalViewMiddlewares:  len(middlewares.Final),
					finalMiddlewares:      len(final),
				},
			},
		},
	}

	method := "GET"
	path := "/"

	for _, test := range tests {
		tt := test

		handlerCounter.viewCalled = false
		handlerCounter.beforeMiddlewares = 0
		handlerCounter.beforeViewMiddlewares = 0
		handlerCounter.afterViewMiddlewares = 0
		handlerCounter.afterMiddlewares = 0
		handlerCounter.finalViewMiddlewares = 0
		handlerCounter.finalMiddlewares = 0

		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			logOutput := &bytes.Buffer{}
			log := log.New(logOutput, "", log.LstdFlags)

			r := newRouter(Config{
				Logger:    log,
				Debug:     true,
				ErrorView: defaultErrorView,
			})
			r.UseBefore(tt.args.before...)
			r.UseAfter(tt.args.after...)
			r.UseFinal(tt.args.final...)
			r.Path(method, path, tt.args.viewFn).Middlewares(tt.args.middlewares)

			ctx := new(fasthttp.RequestCtx)

			h, _ := r.router.Lookup(method, path, ctx)
			h(ctx)

			if ctx.Response.StatusCode() != tt.want.statusCode {
				t.Fatalf("Unexpected status code: '%d', want '%d'", ctx.Response.StatusCode(), tt.want.statusCode)
			}

			if handlerCounter.viewCalled != tt.want.counter.viewCalled {
				t.Errorf("View called = %v, want %v", handlerCounter.viewCalled, tt.want.counter.viewCalled)
			}

			if handlerCounter.beforeMiddlewares != tt.want.counter.beforeMiddlewares {
				t.Errorf("Before middlewares call counter = %v, want %v", handlerCounter.beforeMiddlewares,
					tt.want.counter.beforeMiddlewares)
			}

			if handlerCounter.beforeViewMiddlewares != tt.want.counter.beforeViewMiddlewares {
				t.Errorf("Before view call counter = %v, want %v", handlerCounter.beforeViewMiddlewares,
					tt.want.counter.beforeViewMiddlewares)
			}

			if handlerCounter.afterMiddlewares != tt.want.counter.afterMiddlewares {
				t.Errorf("After middlewares call counter = %v, want %v", handlerCounter.afterMiddlewares,
					tt.want.counter.afterMiddlewares)
			}

			if handlerCounter.afterViewMiddlewares != tt.want.counter.afterViewMiddlewares {
				t.Errorf("After view call counter = %v, want %v", handlerCounter.afterViewMiddlewares,
					tt.want.counter.afterViewMiddlewares)
			}

			if handlerCounter.finalMiddlewares != tt.want.counter.finalMiddlewares {
				t.Errorf("Final middlewares call counter = %v, want %v", handlerCounter.finalMiddlewares,
					tt.want.counter.finalMiddlewares)
			}

			if handlerCounter.finalViewMiddlewares != tt.want.counter.finalViewMiddlewares {
				t.Errorf("Final view call counter = %v, want %v", handlerCounter.finalViewMiddlewares,
					tt.want.counter.finalViewMiddlewares)
			}
		})
	}
}

func TestRouter_handlePath(t *testing.T) {
	r := testRouter()

	path := &Path{
		router:      r,
		method:      fasthttp.MethodGet,
		url:         "/test",
		view:        func(ctx *RequestCtx) error { return nil },
		middlewares: Middlewares{},
		withTimeout: true,
		timeout:     1 * time.Millisecond,
		timeoutMsg:  "Timeout",
		timeoutCode: 404,
	}

	pathOptions := &Path{
		router: r,
		method: fasthttp.MethodOptions,
		url:    "/options",
	}

	paths := []*Path{path, pathOptions}

	for i := range paths {
		p := paths[i]

		r.handlePath(p)

		ctx := new(fasthttp.RequestCtx)

		h, _ := r.router.Lookup(path.method, path.url, ctx)
		if h == nil {
			t.Error("Path is not registered in internal router")
		}

		h, _ = r.router.Lookup(fasthttp.MethodOptions, path.url, ctx)
		if h == nil {
			t.Error("Path (method: OPTIONS) is not registered in internal router")
		}
	}

	// Check if the mutable is disable when adds a new path with the same url of another already added
	for i := range paths {
		p := paths[i]

		r.mutable(true)

		err := catchPanic(func() {
			r.handlePath(p)
		})

		if err == nil {
			t.Errorf("(Method: %s) Panic expected, Mutable must be disabled", p.method)
		}
	}
}

func TestRouter_NewGroupPath(t *testing.T) {
	r := testRouter()
	g := r.NewGroupPath("/fast")

	if g.parent != r {
		t.Errorf("Group router parent == '%p', want %p", g.parent, r)
	}

	if g.prefix != "/fast" {
		t.Errorf("Group router prefix == '%s', want '%s'", g.prefix, "/fast")
	}

	if g.router == nil {
		t.Error("Group router instance is nil")
	}

	if g.routerMutable != r.routerMutable {
		t.Errorf("Group router routerMutable == '%v', want '%v'", g.routerMutable, r.routerMutable)
	}

	if !isEqual(g.errorView, r.errorView) {
		t.Errorf("Group router errorView == '%p', want '%p'", g.errorView, r.errorView)
	}

	if g.handleOPTIONS != r.handleOPTIONS {
		t.Errorf("Group router handleOPTIONS == '%v', want '%v'", g.handleOPTIONS, r.handleOPTIONS)
	}
}

func TestRouter_ListPaths(t *testing.T) {
	server := New(testConfig)

	server.Path("GET", "/foo", func(ctx *RequestCtx) error { return nil })
	server.Path("GET", "/bar", func(ctx *RequestCtx) error { return nil })

	static := server.NewGroupPath("/static")
	static.Static("/buzz", "./docs")

	if !reflect.DeepEqual(server.ListPaths(), server.router.List()) {
		t.Errorf("Router.List() == %v, want %v", server.ListPaths(), server.router.List())
	}
}

func TestRouter_Middlewares(t *testing.T) {
	r := testRouter()
	r.Middlewares(Middlewares{Before: middlewareFns, After: middlewareFns, Skip: middlewareFns})

	if len(r.middlewares.Before) != len(middlewareFns) {
		t.Errorf("Before middlewares are not registered")
	}

	if len(r.middlewares.After) != len(middlewareFns) {
		t.Errorf("After middlewares are not registered")
	}

	if len(r.middlewares.Skip) != len(middlewareFns) {
		t.Errorf("Skip middlewares are not registered")
	}
}

func TestRouter_UseBefore(t *testing.T) {
	r := testRouter()
	r.UseBefore(middlewareFns...)

	if len(r.middlewares.Before) != len(middlewareFns) {
		t.Errorf("Before middlewares are not registered")
	}
}

func TestRouter_UseAfter(t *testing.T) {
	r := testRouter()
	r.UseAfter(middlewareFns...)

	if len(r.middlewares.After) != len(middlewareFns) {
		t.Errorf("After middlewares are not registered")
	}
}

func TestRouter_SkipMiddlewares(t *testing.T) {
	r := testRouter()
	r.SkipMiddlewares(middlewareFns...)

	if len(r.middlewares.Skip) != len(middlewareFns) {
		t.Errorf("Before middlewares are not registered")
	}
}

func TestRouter_Path_Shortcuts(t *testing.T) { //nolint:funlen
	path := "/"
	viewFn := func(ctx *RequestCtx) error { return nil }

	r := testRouter()

	type args struct {
		method string
		fn     func(url string, viewFn View) *Path
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: fasthttp.MethodGet,
			args: args{method: fasthttp.MethodGet, fn: r.GET},
		},
		{
			name: fasthttp.MethodHead,
			args: args{method: fasthttp.MethodHead, fn: r.HEAD},
		},
		{
			name: fasthttp.MethodOptions,
			args: args{method: fasthttp.MethodOptions, fn: r.OPTIONS},
		},
		{
			name: fasthttp.MethodPost,
			args: args{method: fasthttp.MethodPost, fn: r.POST},
		},
		{
			name: fasthttp.MethodPut,
			args: args{method: fasthttp.MethodPut, fn: r.PUT},
		},
		{
			name: fasthttp.MethodPatch,
			args: args{method: fasthttp.MethodPatch, fn: r.PATCH},
		},
		{
			name: fasthttp.MethodDelete,
			args: args{method: fasthttp.MethodDelete, fn: r.DELETE},
		},
		{
			name: fastrouter.MethodWild,
			args: args{method: fastrouter.MethodWild, fn: r.ANY},
		},
	}

	for _, test := range tests {
		test.args.fn(path, viewFn)
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			reqMethod := tt.args.method
			for reqMethod == fastrouter.MethodWild {
				reqMethod = randomHTTPMethod()
			}

			ctx := new(fasthttp.RequestCtx)
			h, _ := r.router.Lookup(reqMethod, path, ctx)

			if h == nil {
				t.Errorf("The path is not registered with method %s", tt.args.method)
			}
		})
	}
}

func TestRouter_Static(t *testing.T) {
	type args struct {
		url      string
		rootPath string
	}

	type want struct {
		routerPath string
	}

	tests := []struct {
		name string
		args args
		want
	}{
		{
			name: "WithoutTrailingSlash",
			args: args{
				url:      "/static",
				rootPath: "/var/www",
			},
			want: want{
				routerPath: "/static/{filepath:*}",
			},
		},
		{
			name: "WithTrailingSlash",
			args: args{
				url:      "/static/",
				rootPath: "/var/www",
			},
			want: want{
				routerPath: "/static/{filepath:*}",
			},
		},
	}

	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			r := testRouter()
			r.Static(tt.args.url, tt.args.rootPath)

			handler, _ := r.router.Lookup("GET", tt.want.routerPath, &fasthttp.RequestCtx{})
			if handler == nil {
				t.Error("Static files is not configured")
			}
		})
	}
}

func TestRouter_StaticCustom(t *testing.T) { //nolint:funlen
	type args struct {
		url      string
		rootPath string
	}

	type want struct {
		routerPath string
	}

	tests := []struct {
		name string
		args args
		want
	}{
		{
			name: "WithoutTrailingSlash",
			args: args{
				url:      "/static",
				rootPath: "./docs",
			},
			want: want{
				routerPath: "/static/{filepath:*}",
			},
		},
		{
			name: "WithTrailingSlash",
			args: args{
				url:      "/static/",
				rootPath: "./docs",
			},
			want: want{
				routerPath: "/static/{filepath:*}",
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			r := testRouter()

			pathRewriteCalled := false

			r.StaticCustom(tt.args.url, &StaticFS{
				Root:               tt.args.rootPath,
				GenerateIndexPages: true,
				AcceptByteRange:    true,
				PathRewrite: func(ctx *RequestCtx) []byte {
					pathRewriteCalled = true

					return ctx.Path()
				},
				PathNotFound: func(ctx *RequestCtx) error {
					return ctx.TextResponse("File not found", 404)
				},
			})

			ctx := new(fasthttp.RequestCtx)
			handler, _ := r.router.Lookup("GET", tt.want.routerPath, ctx)
			if handler == nil {
				t.Fatal("Static files is not configured")
			}

			handler(ctx)

			if !pathRewriteCalled {
				t.Error("Custom path rewrite function is not called")
			}
		})
	}
}

func TestRouter_ServeFile(t *testing.T) {
	type args struct {
		url      string
		filePath string
	}

	filePath := "./README.md"

	body, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	test := struct {
		args args
	}{
		args: args{
			url:      "/readme",
			filePath: filePath,
		},
	}

	r := testRouter()
	r.ServeFile(test.args.url, test.args.filePath)

	ctx := new(fasthttp.RequestCtx)
	handler, _ := r.router.Lookup("GET", test.args.url, ctx)

	if handler == nil {
		t.Error("ServeFile() is not configured")
	}

	handler(ctx)

	if string(ctx.Response.Body()) != string(body) {
		t.Fatal("Invalid response")
	}
}

func TestRouter_Path(t *testing.T) { //nolint:funlen
	type args struct {
		method         string
		url            string
		viewFn         View
		netHTTPHandler http.Handler
		handler        fasthttp.RequestHandler
		timeout        time.Duration
		statusCode     int
	}

	type want struct {
		getPanic bool
	}

	testViewFn := func(ctx *RequestCtx) error {
		if _, err := ctx.WriteString("Test"); err != nil {
			t.Fatalf("Error calling WriteString. %+v", err)
		}

		return nil
	}

	testNetHTTPHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Test")
	}
	testMuxHandler := http.NewServeMux()
	testMuxHandler.HandleFunc("/", testNetHTTPHandler)

	testHandler := func(ctx *fasthttp.RequestCtx) {
		if _, err := ctx.WriteString("Test"); err != nil {
			t.Fatalf("Error in WriteString. %+v", err)
		}
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Path",
			args: args{
				method: "GET",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "TimeoutPath",
			args: args{
				method:  "GET",
				url:     "/",
				viewFn:  testViewFn,
				timeout: 1 * time.Second,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "TimeoutWithCodePath",
			args: args{
				method:     "GET",
				url:        "/",
				viewFn:     testViewFn,
				timeout:    1 * time.Second,
				statusCode: 201,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "NetHTTPPath",
			args: args{
				method:         "GET",
				url:            "/",
				netHTTPHandler: testMuxHandler,
				timeout:        1 * time.Second,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "RequestHandlerPath",
			args: args{
				method:  "GET",
				url:     "/",
				handler: testHandler,
				timeout: 1 * time.Second,
			},
			want: want{
				getPanic: false,
			},
		},
		{
			name: "InvalidMethod",
			args: args{
				method: "get",
				url:    "/",
				viewFn: testViewFn,
			},
			want: want{
				getPanic: true,
			},
		},
		{
			name: "InvalidMethod_Timeout",
			args: args{
				method:  "get",
				url:     "/",
				viewFn:  testViewFn,
				timeout: 1 * time.Second,
			},
			want: want{
				getPanic: true,
			},
		},
	}
	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			defer func() {
				r := recover()

				if tt.want.getPanic && r == nil {
					t.Errorf("Panic expected")
				} else if !tt.want.getPanic && r != nil {
					t.Errorf("Unexpected panic")
				}
			}()

			ctx := new(fasthttp.RequestCtx)
			r := testRouter()

			switch {
			case tt.args.netHTTPHandler != nil:
				r.NetHTTPPath(tt.args.method, tt.args.url, tt.args.netHTTPHandler)
			case tt.args.handler != nil:
				r.RequestHandlerPath(tt.args.method, tt.args.url, tt.args.handler)
			case tt.args.timeout > 0:
				if tt.args.statusCode > 0 {
					r.Path(tt.args.method, tt.args.url, tt.args.viewFn).TimeoutCode(
						tt.args.timeout, "Timeout response message", tt.args.statusCode,
					)
				} else {
					r.Path(
						tt.args.method, tt.args.url, tt.args.viewFn).Timeout(tt.args.timeout, "Timeout response message")
				}
			default:
				r.Path(tt.args.method, tt.args.url, tt.args.viewFn)
			}

			handler, _ := r.router.Lookup("GET", tt.args.url, ctx)
			if handler == nil {
				t.Fatal("Path() is not configured")
			}
		})
	}
}

// Benchmarks.
func Benchmark_handler(b *testing.B) {
	r := testRouter()

	h := r.handler(func(ctx *RequestCtx) error { return nil }, Middlewares{})

	ctx := new(fasthttp.RequestCtx)

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		h(ctx)
	}
}

func Benchmark_RouterHandler(b *testing.B) {
	r := testRouter()
	r.GET("/", func(ctx *RequestCtx) error { return nil })

	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/")

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		r.router.Handler(ctx)
	}
}
