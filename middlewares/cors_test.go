package middlewares

import (
	"github.com/savsgio/atreugo/v10"
	"github.com/valyala/fasthttp"

	"reflect"
	"testing"
)

func TestDefaultCors(t *testing.T) {
	s := DefaultCors()

	if !reflect.DeepEqual(s.allowedOrigins, []string{"*"}) {
		t.Errorf("defaultOptions: AllowedOrigins is not %s", []string{"*"})
	}

	if !reflect.DeepEqual(s.allowedMethods, []string{"GET", "POST"}) {
		t.Errorf("defaultOptions: AllowedMethods is not %s", []string{"GET", "POST"})
	}

	if !reflect.DeepEqual(s.allowedHeaders, []string{"Origin", "Accept", "Content-Type"}) {
		t.Errorf("defaultOptions: AllowedHeaders is not %s", []string{"Origin", "Accept", "Content-Type"})
	}
}

var presetOptions = CorsOptions{
	// if you leave allowedOrigins empty then atreugo will treat it as "*"
	AllowedOrigins: []string{"http://localhost:63342", "192.168.3.1:8000", "APP"},
	// if you leave allowedHeaders empty then atreugo will accept any non-simple headers
	AllowedHeaders: []string{"Content-Type", "content-type"},
	// if you leave this empty, only simple method will be accepted
	AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
	AllowCredentials: true,
	AllowMaxAge:      5600,
}

func TestNewCorsMiddleware(t *testing.T) {
	t.Run("NewCorsMiddleware", func(t *testing.T) {
		s := NewCorsMiddleware(presetOptions)
		if s == nil {
			t.Error("NewCorsMiddleware() CorsOptions error")
			return
		}

		if len(s.allowedOrigins) == 0 {
			ss := defaultOptions.AllowedOrigins
			if ss == nil {
				t.Error("NewCorsMiddleware() allowedOrigins error")
			}
		}

		ctx := new(atreugo.RequestCtx)
		ctx.RequestCtx = new(fasthttp.RequestCtx)
		h := s.handlePreflight(ctx)
		if h != nil {
			t.Error("handlePreflight() error")
		}
	})
}

func TestCorsMiddleware(t *testing.T) {
	type args struct {
		AllowedOrigins []string
		AllowedHeaders []string
		AllowedMethods []string
	}

	type want struct {
		emptyOrigins bool
		emptyHeaders bool
		emptyMethods bool
		err          bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "CorsArray Empty",
			args: args{
				AllowedOrigins: []string{},
				AllowedHeaders: []string{},
				AllowedMethods: []string{},
			},
			want: want{
				emptyOrigins: true,
				emptyHeaders: true,
				emptyMethods: true,
				err:          false,
			},
		},
		{
			name: "CorsArray GT One",
			args: args{
				AllowedOrigins: []string{"http://localhost:63342", "192.168.3.1:8000", "APP"},
				AllowedHeaders: []string{"Content-Type", "content-type"},
				AllowedMethods: []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
			},
			want: want{
				emptyOrigins: false,
				emptyHeaders: false,
				emptyMethods: false,
				err:          false,
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			var customOptions = CorsOptions{
				AllowedOrigins: tt.args.AllowedOrigins,
				AllowedHeaders: tt.args.AllowedHeaders,
				AllowedMethods: tt.args.AllowedMethods,
			}
			newCors := NewCorsMiddleware(customOptions)

			if newCors == nil {
				t.Errorf("%s NewCorsMiddleware() CorsOptions error", "CorsMiddleware")
				return
			}

			ctx := new(atreugo.RequestCtx)
			ctx.RequestCtx = new(fasthttp.RequestCtx)
			err := newCors.CorsMiddleware(ctx)
			if err != nil {
				t.Errorf("CorsMiddleware() error")
			}

			if tt.want.emptyHeaders && (len(newCors.allowedHeaders) == 0) {
				t.Error("Empty allowedHeaders error")
			}

			if tt.want.emptyOrigins && (len(newCors.allowedOrigins) == 0) {
				t.Error("Empty allowedOrigins error")
			}

			if !tt.want.emptyMethods && (len(newCors.allowedMethods) == 0) {
				t.Error("Empty allowedMethods error")
			}

			var isAllowedOrigin = newCors.isAllowedOrigin("APP")

			if tt.want.emptyOrigins && isAllowedOrigin {
				t.Error("isAllowedOrigin() error")
			}

			err = newCors.handlePreflight(ctx)
			if err != nil {
				t.Error(err)
				return
			}

			ctx.Response.Header.Add("Origin", "APP")
			ctx.Response.Header.Add("Access-Control-Allow-Origin", "APP")

			theV := ctx.Response.Header.Peek("Access-Control-Allow-Origin")
			if string(theV) != "APP" {
				t.Error("ctx.Response.Header.Peek error", string(theV))
			}

			ctx.RequestCtx.Response.Header.Add("Vary", "Origin")
			varyHeader := ctx.Response.Header.Peek("Vary")

			if len(varyHeader) == 0 {
				t.Error("varyHeader (great than 0)")
			}

			method := string(ctx.Request.Header.Method())
			if method == fasthttp.MethodOptions {
				if len(newCors.allowedMethods) == 0 {
					t.Error("method pass, allowedMethods error")
				}

				if len(newCors.allowedHeaders) == 0 {
					t.Error("method pass, allowedHeaders error")
				}
			}
		})
	}
}
