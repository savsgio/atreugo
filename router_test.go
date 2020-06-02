package atreugo

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"testing"
	"time"

	fastrouter "github.com/fasthttp/router"
	logger "github.com/savsgio/go-logger/v2"
	"github.com/valyala/fasthttp"
)

var testLog = logger.New("test", "fatal", nil)

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

func randomHTTPMethod() string {
	return httpMethods[rand.Intn(len(httpMethods)-1)]
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
	p1 := &Path{method: fasthttp.MethodGet, url: url}
	p2 := &Path{method: fasthttp.MethodPost, url: url}
	paths := []*Path{p1, p2}

	var errOptionsView error

	headerKey := "key"
	headerValue := "value"
	optionsView := func(ctx *RequestCtx) error {
		ctx.Response.Header.Set(headerKey, headerValue)

		return errOptionsView
	}

	h := buildOptionsView(url, paths, optionsView)

	ctx := AcquireRequestCtx(new(fasthttp.RequestCtx))

	if err := h(ctx); err != errOptionsView {
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

	h = buildOptionsView(url, []*Path{}, optionsView)

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
	r := newRouter(testLog, nil)

	if r.router == nil {
		t.Error("Router instance is nil")
	}

	if r.beginPath != "/" {
		t.Errorf("Router beginPath == '%s', want '%s'", r.beginPath, "/")
	}

	if reflect.ValueOf(r.log).Pointer() != reflect.ValueOf(testLog).Pointer() {
		t.Errorf("Router log == %p, want %p", r.log, testLog)
	}

	if reflect.ValueOf(r.errorView).Pointer() != reflect.ValueOf(defaultErrorView).Pointer() {
		t.Errorf("Router log == %p, want %p", r.errorView, defaultErrorView)
	}
}

func TestRouter_NewGroupPath(t *testing.T) {
	r := newRouter(testLog, nil)
	g := r.NewGroupPath("/fast")

	if g.router == nil {
		t.Error("Group router instance is nil")
	}

	if g.parent != r {
		t.Errorf("Group router parent == '%p', want %p", g.parent, r)
	}

	if g.beginPath != "/fast" {
		t.Errorf("Group router beginPath == '%s', want '%s'", g.beginPath, "/fast")
	}

	if reflect.ValueOf(g.errorView).Pointer() != reflect.ValueOf(r.errorView).Pointer() {
		t.Errorf("Group router log == %p, want %p", g.errorView, r.errorView)
	}

	if reflect.ValueOf(g.log).Pointer() != reflect.ValueOf(r.log).Pointer() {
		t.Errorf("Group router log == %p, want %p", g.log, r.log)
	}
}

func TestRouter_init(t *testing.T) { // nolint:funlen
	var registeredViews []View
	var registeredMiddlewaresView, registeredMiddlewaresOptionsView Middlewares // nolint:wsl

	handlerBuilderViewCalled, handlerBuilderOptionsViewCalled := false, false

	path := &Path{
		handlerBuilder: func(fn View, middle Middlewares) fasthttp.RequestHandler {
			registeredViews = append(registeredViews, fn)
			registeredMiddlewaresView = middle
			handlerBuilderViewCalled = true

			return func(ctx *fasthttp.RequestCtx) {}
		},
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
		handlerBuilder: func(fn View, middle Middlewares) fasthttp.RequestHandler {
			registeredViews = append(registeredViews, fn)
			registeredMiddlewaresOptionsView = middle
			handlerBuilderOptionsViewCalled = true

			return func(ctx *fasthttp.RequestCtx) {}
		},
		method: fasthttp.MethodOptions,
		url:    "/options",
	}

	r := newRouter(testLog, nil)
	r.appendPath(path)
	r.appendPath(pathOptions)

	totalRegisteredViews := len(r.paths) + 1 // Add +1 for auto OPTIONS handle

	r.init()

	// Check if a re-execution raise a panic
	func() {
		defer func() {
			err := recover()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		}()

		r.init()
	}()

	if len(registeredViews) != totalRegisteredViews {
		t.Fatalf("Registered views == %d, want %d", len(registeredViews), totalRegisteredViews)
	}

	ctx := new(fasthttp.RequestCtx)

	h, _ := r.router.Lookup(path.method, path.url, ctx)
	if h == nil {
		t.Error("Path is not registered in internal router")
	}

	h, _ = r.router.Lookup(fasthttp.MethodOptions, path.url, ctx)
	if h == nil {
		t.Error("Path AUTO (method: OPTIONS) is not registered in internal router")
	}

	h, _ = r.router.Lookup(pathOptions.method, pathOptions.url, ctx)
	if h == nil {
		t.Error("Path (method: OPTIONS) is not registered in internal router")
	}

	for _, called := range []bool{handlerBuilderViewCalled, handlerBuilderOptionsViewCalled} {
		if !called {
			t.Error("Path.handlerBuilder is not called")
		}
	}

	if reflect.ValueOf(registeredViews[0]).Pointer() != reflect.ValueOf(path.view).Pointer() {
		t.Errorf("Registered view == %p, want %p", registeredViews[0], path.view)
	}

	for _, middle := range []Middlewares{registeredMiddlewaresView, registeredMiddlewaresOptionsView} {
		if !reflect.DeepEqual(middle, path.middlewares) {
			t.Errorf("Registered middlewares == %v, want %v", middle, &path.middlewares)
		}
	}

	if r.router.GlobalOPTIONS != nil {
		t.Error("GlobalOPTIONS is not nil")
	}

	defer func() {
		err := recover()
		if err == nil {
			t.Error("Panic expected")
		}
	}()

	g := r.NewGroupPath("/test")
	g.init()
}

func TestRouter_buildMiddlewaresChain(t *testing.T) {
	logLevels := []string{"fatal", "debug"}

	mdlws := Middlewares{
		Before: []Middleware{
			func(ctx *RequestCtx) error { return ctx.Next() },
			func(ctx *RequestCtx) error { return ctx.Next() },
		},
		After: []Middleware{func(ctx *RequestCtx) error { return ctx.Next() }},
	}

	for _, level := range logLevels {
		s := New(Config{LogLevel: level})
		s.Middlewares(mdlws)

		chain := s.buildMiddlewaresChain(mdlws.Before[0])

		wantSkipLen := 0
		if len(chain.Skip) != wantSkipLen {
			t.Errorf("Middlewares.Skip length == %d, want %d", len(chain.Skip), wantSkipLen)
		}

		wantBeforeLen := len(mdlws.Before) - 1
		if s.log.DebugEnabled() {
			wantBeforeLen++
		}

		if len(chain.Before) != wantBeforeLen {
			t.Errorf("Middlewares.Before length == %d, want %d", len(chain.Before), wantBeforeLen)
		}

		if s.log.DebugEnabled() && isEqual(chain.Before[0], mdlws.Before[1]) {
			t.Error("First before middleware must be the logger middleware")
		}

		wantAfterLen := len(mdlws.After)
		if len(chain.After) != wantAfterLen {
			t.Errorf("Middlewares.After length == %d, want %d", len(chain.After), wantAfterLen)
		}
	}
}

func TestRouter_handlerExecutionChain(t *testing.T) { //nolint:funlen
	s := New(testAtreugoConfig)

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
	}

	wantOrder := map[string]int{
		"globalBefore": 1,
		"groupBefore":  2,
		"viewBefore":   3,
		"viewAfter":    4,
		"groupAfter":   5,
		"globalAfter":  6,
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
	}).SkipMiddlewares(skipMiddlewareGroup)

	s.init()

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
	r := newRouter(testLog, nil)
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

func TestRouter_AddAndAppendPath(t *testing.T) {
	r := newRouter(testLog, nil)
	foo := r.NewGroupPath("/foo")

	method := fasthttp.MethodOptions
	url := "/"
	fn := func(ctx *RequestCtx) error { return nil }

	foo.addPath(method, url, fn)

	if len(r.paths) == 0 {
		t.Fatal("Path is not added")
	}

	if len(r.customOptionsURLS) == 0 {
		t.Fatal("Custom OPTIONS url is not added")
	}

	p := r.paths[0]

	if reflect.ValueOf(p.handlerBuilder).Pointer() != reflect.ValueOf(foo.handler).Pointer() {
		t.Errorf("Path.view == %p, want %p", p.handlerBuilder, r.handler)
	}

	if p.method != method {
		t.Errorf("Path.method == '%s', want '%s'", p.method, method)
	}

	wantURL := foo.getGroupFullPath(url)
	if p.url != wantURL {
		t.Errorf("Path.url == '%s', want '%s'", p.url, wantURL)
	}

	if r.customOptionsURLS[0] != wantURL {
		t.Errorf("Custom OPTIONS url == %s, want %s", r.customOptionsURLS[0], wantURL)
	}

	if reflect.ValueOf(p.view).Pointer() != reflect.ValueOf(fn).Pointer() {
		t.Errorf("Path.view == %p, want %p", p.view, fn)
	}
}

func TestRouter_handler(t *testing.T) { //nolint:funlen
	type counter struct {
		viewCalled            bool
		beforeMiddlewares     int
		beforeViewMiddlewares int
		afterViewMiddlewares  int
		afterMiddlewares      int
	}

	type args struct {
		viewFn      View
		before      []Middleware
		after       []Middleware
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
				},
			},
		},
		{
			name: "BeforeViewError",
			args: args{
				viewFn: viewFn,
				before: before,
				after:  after,
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
				},
			},
		},
		{
			name: "AfterViewError",
			args: args{
				viewFn: viewFn,
				before: before,
				after:  after,
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

		t.Run(tt.name, func(t *testing.T) {
			logOutput := &bytes.Buffer{}
			log := logger.New(tt.name, "debug", logOutput)

			r := newRouter(log, nil)
			r.UseBefore(tt.args.before...)
			r.UseAfter(tt.args.after...)
			r.Path(method, path, tt.args.viewFn).Middlewares(tt.args.middlewares)
			r.init()

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

			if logOutput.Len() == 0 {
				t.Errorf("Debug trace has not been write in log when logger 'debug' is enabled")
			}
		})
	}
}

func TestRouter_Middlewares(t *testing.T) {
	r := newRouter(testLog, nil)
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
	r := newRouter(testLog, nil)
	r.UseBefore(middlewareFns...)

	if len(r.middlewares.Before) != len(middlewareFns) {
		t.Errorf("Before middlewares are not registered")
	}
}

func TestRouter_UseAfter(t *testing.T) {
	r := newRouter(testLog, nil)
	r.UseAfter(middlewareFns...)

	if len(r.middlewares.After) != len(middlewareFns) {
		t.Errorf("After middlewares are not registered")
	}
}

func TestRouter_SkipMiddlewares(t *testing.T) {
	r := newRouter(testLog, nil)
	r.SkipMiddlewares(middlewareFns...)

	if len(r.middlewares.Skip) != len(middlewareFns) {
		t.Errorf("Before middlewares are not registered")
	}
}

func TestRouter_Path_Shortcuts(t *testing.T) { //nolint:funlen
	path := "/"
	viewFn := func(ctx *RequestCtx) error { return nil }

	r := newRouter(testLog, nil)

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

	r.init()

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
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
			defer func() {
				r := recover()

				if tt.want.getPanic && r == nil {
					t.Errorf("Panic expected")
				} else if !tt.want.getPanic && r != nil {
					t.Errorf("Unexpected panic")
				}
			}()

			ctx := new(fasthttp.RequestCtx)
			r := newRouter(testLog, nil)

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

			r.init()

			handler, _ := r.router.Lookup("GET", tt.args.url, ctx)
			if handler == nil {
				t.Fatal("Path() is not configured")
			}

			// ctx.Request.SetRequestURI(tt.args.url)

			// ctx.Request.Header.SetMethod(tt.args.method)
			// handler(ctx)

			// if string(ctx.Response.Body()) != "Test" {
			// 	t.Error("Error")
			// }
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
			r := newRouter(testLog, nil)
			r.Static(tt.args.url, tt.args.rootPath)
			r.init()

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
			r := newRouter(testLog, nil)

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
			r.init()

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
	body, err := ioutil.ReadFile(filePath)

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

	r := newRouter(testLog, nil)
	r.ServeFile(test.args.url, test.args.filePath)
	r.init()

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

func TestRouter_ListPaths(t *testing.T) {
	server := New(testAtreugoConfig)

	server.Path("GET", "/foo", func(ctx *RequestCtx) error { return nil })
	server.Path("GET", "/bar", func(ctx *RequestCtx) error { return nil })

	static := server.NewGroupPath("/static")
	static.Static("/buzz", "./docs")

	if !reflect.DeepEqual(server.ListPaths(), server.router.List()) {
		t.Errorf("Router.List() == %v, want %v", server.ListPaths(), server.router.List())
	}
}

// Benchmarks.
func Benchmark_handler(b *testing.B) {
	r := newRouter(testLog, nil)

	h := r.handler(func(ctx *RequestCtx) error { return nil }, Middlewares{})

	ctx := new(fasthttp.RequestCtx)

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		h(ctx)
	}
}

func Benchmark_RouterHandler(b *testing.B) {
	r := newRouter(testLog, nil)
	r.GET("/", func(ctx *RequestCtx) error { return nil })
	r.init()

	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/")

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		r.router.Handler(ctx)
	}
}
