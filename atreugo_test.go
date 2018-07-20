package atreugo

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/erikdubbelboer/fasthttp"
)

func TestAtreugoServer(t *testing.T) {
	type args struct {
		viewFn        View
		middlewareFns []Middleware
	}
	type want struct {
		statusCode        int
		viewCalled        bool
		middleWareCounter int
	}

	viewCalled := false
	middleWareCounter := 0

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "AllOk",
			args: args{
				viewFn: func(ctx *fasthttp.RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *fasthttp.RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
				},
			},
			want: want{
				statusCode:        fasthttp.StatusOK,
				viewCalled:        true,
				middleWareCounter: 1,
			},
		},
		{
			name: "FirstMiddlewareError",
			args: args{
				viewFn: func(ctx *fasthttp.RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *fasthttp.RequestCtx) (int, error) {
						return 403, errors.New("Bad request")
					},
					func(ctx *fasthttp.RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
				},
			},
			want: want{
				statusCode:        403,
				viewCalled:        false,
				middleWareCounter: 0,
			},
		},
		{
			name: "SecondMiddlewareError",
			args: args{
				viewFn: func(ctx *fasthttp.RequestCtx) error {
					viewCalled = true
					return nil
				},
				middlewareFns: []Middleware{
					func(ctx *fasthttp.RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
					func(ctx *fasthttp.RequestCtx) (int, error) {
						return 403, errors.New("Bad request")
					},
				},
			},
			want: want{
				statusCode:        403,
				viewCalled:        false,
				middleWareCounter: 1,
			},
		},
		{
			name: "ViewError",
			args: args{
				viewFn: func(ctx *fasthttp.RequestCtx) error {
					viewCalled = true
					return errors.New("Fake error")
				},
				middlewareFns: []Middleware{
					func(ctx *fasthttp.RequestCtx) (int, error) {
						middleWareCounter++
						return 0, nil
					},
				},
			},
			want: want{
				statusCode:        fasthttp.StatusInternalServerError,
				viewCalled:        true,
				middleWareCounter: 1,
			},
		},
	}

	for _, tt := range tests {
		viewCalled = false
		middleWareCounter = 0

		t.Run(tt.name, func(t *testing.T) {
			ctx := new(fasthttp.RequestCtx)

			s := New(&Config{})
			s.UseMiddleware(tt.args.middlewareFns...)

			s.handler(tt.args.viewFn)(ctx)

			if viewCalled != tt.want.viewCalled {
				t.Errorf("View called = %v, want %v", viewCalled, tt.want.viewCalled)
			}

			if middleWareCounter != tt.want.middleWareCounter {
				t.Errorf("Middleware call counter = %v, want %v", middleWareCounter, tt.want.middleWareCounter)
			}

			responseStatusCode := ctx.Response.StatusCode()
			if responseStatusCode != tt.want.statusCode {
				t.Errorf("Status code = %v, want %v", responseStatusCode, tt.want.statusCode)
			}
		})
	}
}

func TestAtreugo_getListener(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
		want net.Listener
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.getListener(tt.args.addr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Atreugo.getListener() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtreugo_serve(t *testing.T) {
	type args struct {
		ln       net.Listener
		protocol string
		addr     string
	}
	tests := []struct {
		name    string
		s       *Atreugo
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.serve(tt.args.ln, tt.args.protocol, tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("Atreugo.serve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAtreugo_serveGracefully(t *testing.T) {
	type args struct {
		ln       net.Listener
		protocol string
		addr     string
	}
	tests := []struct {
		name    string
		s       *Atreugo
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.serveGracefully(tt.args.ln, tt.args.protocol, tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("Atreugo.serveGracefully() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAtreugo_Static(t *testing.T) {
	type args struct {
		rootStaticDirPath string
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Static(tt.args.rootStaticDirPath)
		})
	}
}

func TestAtreugo_Path(t *testing.T) {
	type args struct {
		httpMethod string
		url        string
		viewFn     View
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Path(tt.args.httpMethod, tt.args.url, tt.args.viewFn)
		})
	}
}

func TestAtreugo_UseMiddleware(t *testing.T) {
	type args struct {
		fns []Middleware
	}
	tests := []struct {
		name string
		s    *Atreugo
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.UseMiddleware(tt.args.fns...)
		})
	}
}

func TestAtreugo_ListenAndServe(t *testing.T) {
	tests := []struct {
		name    string
		s       *Atreugo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.ListenAndServe(); (err != nil) != tt.wantErr {
				t.Errorf("Atreugo.ListenAndServe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
