package atreugo

import (
	"errors"
	"testing"

	"github.com/valyala/fasthttp"
)

func Test_execute(t *testing.T) { //nolint:funlen
	type args struct {
		hs []Middleware
	}

	type want struct {
		statusCode int
		counter    int
		err        bool
	}

	counter := 0

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "OK",
			args: args{
				hs: []Middleware{
					func(ctx *RequestCtx) error {
						counter++
						return ctx.Next()
					},
					func(ctx *RequestCtx) error {
						counter++
						return ctx.Next()
					},
				},
			},
			want: want{
				statusCode: fasthttp.StatusOK,
				counter:    2,
				err:        false,
			},
		},
		{
			name: "Response",
			args: args{
				hs: []Middleware{
					func(ctx *RequestCtx) error {
						counter++
						return ctx.TextResponse("Go")
					},
					func(ctx *RequestCtx) error {
						counter++
						return ctx.Next()
					},
				},
			},
			want: want{
				statusCode: fasthttp.StatusOK,
				counter:    1,
				err:        false,
			},
		},
		{
			name: "Error",
			args: args{
				hs: []Middleware{
					func(ctx *RequestCtx) error {
						counter++
						return errors.New("middleware error")
					},
					func(ctx *RequestCtx) error {
						counter++
						return ctx.Next()
					},
				},
			},
			want: want{
				statusCode: fasthttp.StatusInternalServerError,
				counter:    1,
				err:        true,
			},
		},
	}
	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			counter = 0
			ctx := acquireRequestCtx(new(fasthttp.RequestCtx))

			err := execute(ctx, tt.args.hs)
			if (err != nil) != tt.want.err {
				t.Errorf("execute() unexpected error: %v", err)
			} else if counter != tt.want.counter {
				t.Errorf("execute() counter = %d, want %d", counter, tt.want.counter)
			}
		})
	}
}

func Test_viewToHandler(t *testing.T) {
	called := false
	err := errors.New("error")

	view := func(ctx *RequestCtx) error {
		called = true
		return err
	}

	ctx := new(fasthttp.RequestCtx)

	handler := viewToHandler(view, defaultErrorView)
	handler(ctx)

	if !called {
		t.Error("View is not called")
	}

	if ctx.Response.StatusCode() != fasthttp.StatusInternalServerError {
		t.Errorf("Status code == %d, want %d", ctx.Response.StatusCode(), fasthttp.StatusInternalServerError)
	}

	if string(ctx.Response.Body()) != err.Error() {
		t.Errorf("Response body == %s, want %s", ctx.Response.Body(), err.Error())
	}
}
