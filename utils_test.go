package atreugo

import (
	"errors"
	"testing"
)

func Test_panicOnError(t *testing.T) {
	type args struct {
		err  error
		want bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Panic",
			args: args{
				err:  errors.New("TestPanic"),
				want: true,
			},
		},
		{
			name: "NotPanic",
			args: args{
				err:  nil,
				want: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()

				if tt.args.want && r == nil {
					t.Errorf("panicOnError(): '%v', want '%v'", false, tt.args.want)
				} else if !tt.args.want && r != nil {
					t.Errorf("panicOnError(): '%v', want '%v'", true, tt.args.want)
				}
			}()

			panicOnError(tt.args.err)
		})
	}
}

func Test_indexOf(t *testing.T) {
	type args struct {
		vs []string
		t  string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Found",
			args: args{
				vs: []string{"savsgio", "development", "Atreugo"},
				t:  "Atreugo",
			},
			want: 2,
		},
		{
			name: "NotFound",
			args: args{
				vs: []string{"savsgio", "development", "Atreugo"},
				t:  "Yeah",
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := indexOf(tt.args.vs, tt.args.t); got != tt.want {
				t.Errorf("indexOf(): '%v', want: '%v'", got, tt.want)
			}
		})
	}
}

func Test_include(t *testing.T) {
	type args struct {
		vs []string
		t  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Found",
			args: args{
				vs: []string{"savsgio", "development", "Atreugo"},
				t:  "Atreugo",
			},
			want: true,
		},
		{
			name: "NotFound",
			args: args{
				vs: []string{"savsgio", "development", "Atreugo"},
				t:  "Yeah",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := include(tt.args.vs, tt.args.t); got != tt.want {
				t.Errorf("include() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_execMiddlewares(t *testing.T) {
	type args struct {
		middlewares []Middleware
	}
	type want struct {
		statusCode int
		err        bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "OK",
			args: args{
				middlewares: []Middleware{
					func(ctx *RequestCtx) (int, error) {
						return 200, nil
					},
				},
			},
			want: want{
				statusCode: 200,
				err:        false,
			},
		},
		{
			name: "Error",
			args: args{
				middlewares: []Middleware{
					func(ctx *RequestCtx) (int, error) {
						return 500, errors.New("middleware error")
					},
				},
			},
			want: want{
				statusCode: 500,
				err:        true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := new(RequestCtx)

			statusCode, err := execMiddlewares(ctx, tt.args.middlewares)
			if (err != nil) != tt.want.err {
				t.Errorf("execMiddlewares() unexpected error: %v", err)
			} else if statusCode != tt.want.statusCode {
				t.Errorf("execMiddlewares() status code = %v, want %v", statusCode, tt.want)
			}
		})
	}
}
