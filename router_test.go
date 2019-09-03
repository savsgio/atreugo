package atreugo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	logger "github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
)

var testLog = logger.New("test", "fatal", nil)

func TestRouter_newRouter(t *testing.T) {
	r := newRouter(testLog)

	if reflect.ValueOf(r.log).Pointer() != reflect.ValueOf(testLog).Pointer() {
		t.Errorf("Router log == %p, want %p", r.log, testLog)
	}

	if r.router == nil {
		t.Error("Router instance is nil")
	}
}

func TestRouter_NewGroupPath(t *testing.T) {
	r := newRouter(testLog)
	g := r.NewGroupPath("/fast")

	if reflect.ValueOf(g.log).Pointer() != reflect.ValueOf(r.log).Pointer() {
		t.Errorf("Group log == %p, want %p", g.log, r.log)
	}

	if g.router == nil {
		t.Error("Group router instance is nil")
	}
}

func TestRouter_middlewares(t *testing.T) {
	s := New(testAtreugoConfig)

	method := "GET"
	url := "/foo"

	viewCalled := false

	callOrder := map[string]int{
		"globalBefore": 0,
		"groupBefore":  0,
		"filterBefore": 0,
		"filterAfter":  0,
		"groupAfter":   0,
		"globalAfter":  0,
	}

	wantOrder := map[string]int{
		"globalBefore": 1,
		"groupBefore":  2,
		"filterBefore": 3,
		"filterAfter":  4,
		"groupAfter":   5,
		"globalAfter":  6,
	}

	index := 0

	s.UseBefore(func(ctx *RequestCtx) (int, error) {
		index++
		callOrder["globalBefore"] = index
		return 0, nil
	})
	s.UseAfter(func(ctx *RequestCtx) (int, error) {
		index++
		callOrder["globalAfter"] = index
		return 0, nil
	})

	v1 := s.NewGroupPath("/v1")
	v1.UseBefore(func(ctx *RequestCtx) (int, error) {
		index++
		callOrder["groupBefore"] = index
		return 0, nil
	})
	v1.UseAfter(func(ctx *RequestCtx) (int, error) {
		index++
		callOrder["groupAfter"] = index
		return 0, nil
	})

	filters := Filters{
		Before: []Middleware{func(ctx *RequestCtx) (int, error) {
			index++
			callOrder["filterBefore"] = index
			return 0, nil
		}},
		After: []Middleware{func(ctx *RequestCtx) (int, error) {
			index++
			callOrder["filterAfter"] = index
			return 0, nil
		}},
	}

	v1.PathWithFilters(method, url, func(ctx *RequestCtx) error {
		viewCalled = true
		return nil
	}, filters)

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

}

func TestRouter_middlewareStopAfter(t *testing.T) {
	s := New(testAtreugoConfig)
	s.UseAfter(func(ctx *RequestCtx) (int, error) {
		return 0, nil
	})
	s.UseAfter(func(ctx *RequestCtx) (int, error) {
		return 203, nil
	})
	s.UseAfter(func(ctx *RequestCtx) (int, error) {
		return 403, errors.New("this middleware shouldn't be reached")
	})

	s.PathWithFilters("GET", "/", func(ctx *RequestCtx) error {
		ctx.Response.SetStatusCode(500)
		return nil
	}, Filters{})

	ctx := new(fasthttp.RequestCtx)
	h, _ := s.router.Lookup("GET", "/", ctx)
	if h == nil {
		t.Fatal("Registered handler is nil")
	}
	h(ctx)

	if status := ctx.Response.StatusCode(); status != 203 {
		t.Errorf("Expected status 203, but received %v", status)
	}
}

func TestRouter_middlewareStopBefore_statusCode(t *testing.T) {
	s := New(testAtreugoConfig)
	s.UseBefore(func(ctx *RequestCtx) (int, error) {
		return 0, nil
	})
	s.UseBefore(func(ctx *RequestCtx) (int, error) {
		return 203, nil
	})
	s.UseBefore(func(ctx *RequestCtx) (int, error) {
		return 403, errors.New("this middleware shouldn't be reached")
	})

	s.PathWithFilters("GET", "/", func(ctx *RequestCtx) error {
		ctx.Response.SetStatusCode(500)
		return errors.New("the handler shouldn't be reached")
	}, Filters{})

	ctx := new(fasthttp.RequestCtx)
	h, _ := s.router.Lookup("GET", "/", ctx)
	if h == nil {
		t.Fatal("Registered handler is nil")
	}
	h(ctx)

	if status := ctx.Response.StatusCode(); status != 203 {
		t.Errorf("Expected status 203, but received %v", status)
	}
}

func TestRouter_middlewareStopBefore_error(t *testing.T) {
	s := New(testAtreugoConfig)
	s.UseBefore(func(ctx *RequestCtx) (int, error) {
		return 0, nil
	})
	s.UseBefore(func(ctx *RequestCtx) (int, error) {
		return 0, errors.New("stop here!")
	})
	s.UseBefore(func(ctx *RequestCtx) (int, error) {
		return 403, errors.New("this middleware shouldn't be reached")
	})

	s.PathWithFilters("GET", "/", func(ctx *RequestCtx) error {
		ctx.Response.SetStatusCode(203)
		return errors.New("the handler shouldn't be reached")
	}, Filters{})

	ctx := new(fasthttp.RequestCtx)
	h, _ := s.router.Lookup("GET", "/", ctx)
	if h == nil {
		t.Fatal("Registered handler is nil")
	}
	h(ctx)

	if status := ctx.Response.StatusCode(); status != 500 {
		t.Errorf("Expected status 500, but received %v", status)
	}
}

func TestRouter_getGroupFullPath(t *testing.T) {
	r := newRouter(testLog)
	foo := r.NewGroupPath("/foo")
	bar := foo.NewGroupPath("/bar")
	buz := bar.NewGroupPath("/buz")

	path := "/atreugo"

	fullPath := buz.getGroupFullPath(path)
	expected := "/foo/bar/buz/atreugo"
	if fullPath != expected {
		t.Errorf("Router.getGroupFullPath == %s, want %s", fullPath, expected)
	}

	fullPath = bar.getGroupFullPath(path)
	expected = "/foo/bar/atreugo"
	if fullPath != expected {
		t.Errorf("Router.getGroupFullPath == %s, want %s", fullPath, expected)
	}

	fullPath = foo.getGroupFullPath(path)
	expected = "/foo/atreugo"
	if fullPath != expected {
		t.Errorf("Router.getGroupFullPath == %s, want %s", fullPath, expected)
	}

}

func TestRouter_handler(t *testing.T) {
	type counter struct {
		viewCalled        bool
		beforeMiddlewares int
		beforeFilters     int
		afterFilters      int
		afterMiddlewares  int
	}

	type args struct {
		viewFn  View
		before  []Middleware
		after   []Middleware
		filters Filters
	}
	type want struct {
		statusCode int
		counter    counter
	}

	handlerCounter := counter{}
	err := errors.New("test error")

	viewFn := func(ctx *RequestCtx) error {
		handlerCounter.viewCalled = true
		return nil
	}
	before := []Middleware{
		func(ctx *RequestCtx) (int, error) {
			handlerCounter.beforeMiddlewares++
			return 0, nil
		},
	}
	after := []Middleware{
		func(ctx *RequestCtx) (int, error) {
			handlerCounter.afterMiddlewares++
			return 0, nil
		},
	}
	filters := Filters{
		Before: []Middleware{
			func(ctx *RequestCtx) (int, error) {
				handlerCounter.beforeFilters++
				return 0, nil
			},
		},
		After: []Middleware{
			func(ctx *RequestCtx) (int, error) {
				handlerCounter.afterFilters++
				return 0, nil
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
				viewFn:  viewFn,
				before:  before,
				after:   after,
				filters: filters,
			},
			want: want{
				statusCode: fasthttp.StatusOK,
				counter: counter{
					viewCalled:        true,
					beforeMiddlewares: len(before),
					beforeFilters:     len(filters.Before),
					afterFilters:      len(filters.After),
					afterMiddlewares:  len(after),
				},
			},
		},
		{
			name: "ViewError",
			args: args{
				viewFn: func(ctx *RequestCtx) error {
					return err
				},
				before:  before,
				after:   after,
				filters: filters,
			},
			want: want{
				statusCode: fasthttp.StatusInternalServerError,
				counter: counter{
					viewCalled:        false,
					beforeMiddlewares: len(before),
					beforeFilters:     len(filters.Before),
					afterFilters:      0,
					afterMiddlewares:  0,
				},
			},
		},
		{
			name: "BeforeMiddlewaresError",
			args: args{
				viewFn: viewFn,
				before: []Middleware{
					func(ctx *RequestCtx) (int, error) {
						handlerCounter.beforeMiddlewares++
						return fasthttp.StatusBadRequest, err
					},
				},
				after:   after,
				filters: filters,
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
				counter: counter{
					viewCalled:        false,
					beforeMiddlewares: 1,
					beforeFilters:     0,
					afterFilters:      0,
					afterMiddlewares:  0,
				},
			},
		},
		{
			name: "BeforeFiltersError",
			args: args{
				viewFn: viewFn,
				before: before,
				after:  after,
				filters: Filters{
					Before: []Middleware{
						func(ctx *RequestCtx) (int, error) {
							handlerCounter.beforeFilters++
							return fasthttp.StatusBadRequest, err
						},
					},
					After: []Middleware{
						func(ctx *RequestCtx) (int, error) {
							handlerCounter.afterFilters++
							return 0, nil
						},
					},
				},
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
				counter: counter{
					viewCalled:        false,
					beforeMiddlewares: len(before),
					beforeFilters:     1,
					afterFilters:      0,
					afterMiddlewares:  0,
				},
			},
		},
		{
			name: "AfterFiltersError",
			args: args{
				viewFn: viewFn,
				before: before,
				after:  after,
				filters: Filters{
					Before: []Middleware{
						func(ctx *RequestCtx) (int, error) {
							handlerCounter.beforeFilters++
							return 0, nil
						},
					},
					After: []Middleware{
						func(ctx *RequestCtx) (int, error) {
							handlerCounter.afterFilters++
							return fasthttp.StatusBadRequest, err
						},
					},
				},
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
				counter: counter{
					viewCalled:        true,
					beforeMiddlewares: len(before),
					beforeFilters:     1,
					afterFilters:      1,
					afterMiddlewares:  0,
				},
			},
		},
		{
			name: "AfterMiddlewaresError",
			args: args{
				viewFn: viewFn,
				before: before,
				after: []Middleware{
					func(ctx *RequestCtx) (int, error) {
						handlerCounter.afterMiddlewares++
						return fasthttp.StatusBadRequest, err
					},
				},
				filters: filters,
			},
			want: want{
				statusCode: fasthttp.StatusBadRequest,
				counter: counter{
					viewCalled:        true,
					beforeMiddlewares: len(before),
					beforeFilters:     len(filters.Before),
					afterFilters:      len(filters.After),
					afterMiddlewares:  1,
				},
			},
		},
	}

	httpMethod := "GET"
	path := "/"

	for _, tt := range tests {
		handlerCounter.viewCalled = false
		handlerCounter.beforeMiddlewares = 0
		handlerCounter.beforeFilters = 0
		handlerCounter.afterFilters = 0
		handlerCounter.afterMiddlewares = 0

		t.Run(tt.name, func(t *testing.T) {
			r := newRouter(testLog)
			r.UseBefore(tt.args.before...)
			r.UseAfter(tt.args.after...)
			r.PathWithFilters(httpMethod, path, tt.args.viewFn, tt.args.filters)

			ctx := new(fasthttp.RequestCtx)

			h, _ := r.router.Lookup(httpMethod, path, ctx)
			h(ctx)

			if ctx.Response.StatusCode() != tt.want.statusCode {
				t.Fatalf("[test:%s] Unexpected status code: '%d', want '%d'", tt.name, ctx.Response.StatusCode(), tt.want.statusCode)
			}

			if handlerCounter.viewCalled != tt.want.counter.viewCalled {
				t.Errorf("[test:%s] View called = %v, want %v", tt.name, handlerCounter.viewCalled, tt.want.counter.viewCalled)
			}

			if handlerCounter.beforeMiddlewares != tt.want.counter.beforeMiddlewares {
				t.Errorf("[test:%s] Before middlewares call counter = %v, want %v", tt.name, handlerCounter.beforeMiddlewares,
					tt.want.counter.beforeMiddlewares)
			}

			if handlerCounter.beforeFilters != tt.want.counter.beforeFilters {
				t.Errorf("[test:%s] Before filters call counter = %v, want %v", tt.name, handlerCounter.beforeFilters,
					tt.want.counter.beforeFilters)
			}

			if handlerCounter.afterMiddlewares != tt.want.counter.afterMiddlewares {
				t.Errorf("[test:%s] After middlewares call counter = %v, want %v", tt.name, handlerCounter.afterMiddlewares,
					tt.want.counter.afterMiddlewares)
			}

			if handlerCounter.afterFilters != tt.want.counter.afterFilters {
				t.Errorf("[test:%s] After filters call counter = %v, want %v", tt.name, handlerCounter.afterFilters,
					tt.want.counter.afterFilters)
			}

		})
	}
}

func TestRouter_UseBefore(t *testing.T) {
	middlewareFns := []Middleware{
		func(ctx *RequestCtx) (int, error) {
			return 403, errors.New("Bad request")
		},
		func(ctx *RequestCtx) (int, error) {
			return 0, nil
		},
	}

	r := newRouter(testLog)
	r.UseBefore(middlewareFns...)

	if len(r.beforeMiddlewares) != len(middlewareFns) {
		t.Errorf("Middlewares are not registered")
	}
}

func TestRouter_UseAfter(t *testing.T) {
	middlewareFns := []Middleware{
		func(ctx *RequestCtx) (int, error) {
			return 403, errors.New("Bad request")
		},
		func(ctx *RequestCtx) (int, error) {
			return 0, nil
		},
	}

	r := newRouter(testLog)
	r.UseAfter(middlewareFns...)

	if len(r.afterMiddlewares) != len(middlewareFns) {
		t.Errorf("Middlewares are not registered")
	}
}

func TestRouter_Path(t *testing.T) {
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
		ctx.WriteString("Test")
		return nil
	}

	testNetHTTPHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Test")
	}
	testMuxHandler := http.NewServeMux()
	testMuxHandler.HandleFunc("/", testNetHTTPHandler)

	testHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("Test")
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
	for _, tt := range tests {
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

			r := newRouter(testLog)
			if tt.args.netHTTPHandler != nil {
				r.NetHTTPPath(tt.args.method, tt.args.url, tt.args.netHTTPHandler)
			} else if tt.args.handler != nil {
				r.RequestHandlerPath(tt.args.method, tt.args.url, tt.args.handler)
			} else if tt.args.timeout > 0 {
				if tt.args.statusCode > 0 {
					r.TimeoutWithCodePath(
						tt.args.method, tt.args.url, tt.args.viewFn, tt.args.timeout, "Timeout response message",
						tt.args.statusCode,
					)
				} else {
					r.TimeoutPath(
						tt.args.method, tt.args.url, tt.args.viewFn, tt.args.timeout, "Timeout response message",
					)
				}
			} else {
				r.Path(tt.args.method, tt.args.url, tt.args.viewFn)
			}

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
				routerPath: "/static/*filepath",
			},
		},
		{
			name: "WithTrailingSlash",
			args: args{
				url:      "/static/",
				rootPath: "/var/www",
			},
			want: want{
				routerPath: "/static/*filepath",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter(testLog)
			r.Static(tt.args.url, tt.args.rootPath)

			handler, _ := r.router.Lookup("GET", tt.want.routerPath, &fasthttp.RequestCtx{})
			if handler == nil {
				t.Error("Static files is not configured")
			}
		})
	}
}

func TestRouter_StaticCustom(t *testing.T) {
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
				routerPath: "/static/*filepath",
			},
		},
		{
			name: "WithTrailingSlash",
			args: args{
				url:      "/static/",
				rootPath: "./docs",
			},
			want: want{
				routerPath: "/static/*filepath",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter(testLog)

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

			handler, _ := r.router.Lookup("GET", tt.want.routerPath, &fasthttp.RequestCtx{})
			if handler == nil {
				t.Fatal("Static files is not configured")
			}

			ctx := new(fasthttp.RequestCtx)
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

	r := newRouter(testLog)
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

// Benchmarks
func Benchmark_handler(b *testing.B) {
	r := newRouter(testLog)
	viewFn := func(ctx *RequestCtx) error {
		return ctx.HTTPResponse("Hello world")
	}
	ctx := new(fasthttp.RequestCtx)
	h := r.handler(viewFn, emptyFilters)

	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		h(ctx)
	}
}
