package middlewares

import (
	"testing"

	"github.com/savsgio/atreugo/v10"
	"github.com/valyala/fasthttp"
)

func Test_NewCorsMiddleware(t *testing.T) { //nolint:funlen
	type args struct {
		method string
		origin string
		vary   string
		cors   CORS
	}

	type want struct {
		origin           string
		vary             string
		allowedHeaders   string
		allowedMethods   string
		exposedHeaders   string
		allowCredentials string
		allowMaxAge      string
	}

	tests := []struct {
		args args
		want want
	}{
		{
			args: args{
				method: "OPTIONS",
				vary:   "Accept-Encoding",
				origin: "https://cors.test",
				cors: CORS{
					AllowedOrigins:   []string{"https://other.domain.test", "https://cors.test"},
					AllowedHeaders:   []string{"Content-Type", "X-Custom"},
					AllowedMethods:   []string{"GET", "POST", "DELETE"},
					ExposedHeaders:   []string{"Content-Length", "Authorization"},
					AllowCredentials: true,
					AllowMaxAge:      5600,
				},
			},
			want: want{
				origin:           "https://cors.test",
				vary:             "Accept-Encoding, Origin",
				allowedHeaders:   "Content-Type, X-Custom",
				allowedMethods:   "GET, POST, DELETE",
				exposedHeaders:   "Content-Length, Authorization",
				allowCredentials: "true",
				allowMaxAge:      "5600",
			},
		},
		{
			args: args{
				method: "POST",
				vary:   "",
				origin: "https://cors.test",
				cors: CORS{
					AllowedOrigins:   []string{"*"},
					AllowedHeaders:   []string{"Content-Type", "X-Custom"},
					AllowedMethods:   []string{"GET", "POST", "DELETE"},
					ExposedHeaders:   []string{"Content-Length", "Authorization"},
					AllowCredentials: true,
					AllowMaxAge:      5600,
				},
			},
			want: want{
				origin:           "https://cors.test",
				vary:             "Origin",
				allowedHeaders:   "",
				allowedMethods:   "",
				exposedHeaders:   "Content-Length, Authorization",
				allowCredentials: "true",
				allowMaxAge:      "",
			},
		},
		{
			args: args{
				method: "POST",
				vary:   "",
				origin: "https://cors.test",
				cors: CORS{
					AllowedOrigins:   []string{"https://other.domain.test"},
					AllowedHeaders:   []string{"Content-Type", "X-Custom"},
					AllowedMethods:   []string{"GET", "POST", "DELETE"},
					ExposedHeaders:   []string{"Content-Length", "Authorization"},
					AllowCredentials: true,
					AllowMaxAge:      5600,
				},
			},
			want: want{
				origin:           "",
				vary:             "",
				allowedHeaders:   "",
				allowedMethods:   "",
				exposedHeaders:   "",
				allowCredentials: "",
				allowMaxAge:      "",
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run("", func(t *testing.T) {
			m := NewCORSMiddleware(tt.args.cors)

			ctx := new(atreugo.RequestCtx)
			ctx.RequestCtx = new(fasthttp.RequestCtx)

			ctx.Request.Header.Set(originHeaderName, tt.want.origin)
			ctx.Request.Header.SetMethod(tt.args.method)
			ctx.Response.Header.Set(varyHeaderName, tt.args.vary)

			if err := m(ctx); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			wantHeadersValue := map[string]string{
				varyHeaderName:             tt.want.vary,
				allowOriginHeaderName:      tt.want.origin,
				allowCredentialsHeaderName: tt.want.allowCredentials,
				allowHeadersHeaderName:     tt.want.allowedHeaders,
				allowMethodsHeaderName:     tt.want.allowedMethods,
				exposeHeadersHeaderName:    tt.want.exposedHeaders,
				maxAgeHeaderName:           tt.want.allowMaxAge,
			}

			for headerName, want := range wantHeadersValue {
				got := string(ctx.Response.Header.Peek(headerName))
				if got != want {
					t.Errorf("Header: %s == %s, want %s", headerName, got, want)
				}
			}
		})
	}
}

func Test_cors_isAllowedOrigin(t *testing.T) {
	allowedOrigins := []string{"https://other.domain.test", "https://cors.test"}

	origin := allowedOrigins[0]
	if allowed := isAllowedOrigin(allowedOrigins, origin); !allowed {
		t.Errorf("Origin == %s, must be allowed", origin)
	}

	origin = "other"

	if allowed := isAllowedOrigin(allowedOrigins, origin); allowed {
		t.Errorf("Origin == %s, must not be allowed", origin)
	}

	allowedOrigins = []string{"*"}

	if allowed := isAllowedOrigin(allowedOrigins, origin); !allowed {
		t.Errorf("Origin == %s, must be allowed", origin)
	}
}
