package middlewares

import (
	"github.com/savsgio/atreugo/v10"
	"github.com/valyala/fasthttp"

	"testing"
)

var presetOptions = CorsOptions{
	AllowedOrigins:   []string{"http://localhost:63342", "192.168.3.1:8000", "APP"},
	AllowedHeaders:   []string{"Content-Type", "content-type"},
	AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
	ExposedHeaders:   []string{"Content-Length, Authorization"},
	AllowedVary:      []string{"Origin, User-Agent"},
	AllowCredentials: true,
	AllowMaxAge:      5600,
}

func TestNewCorsMiddleware(t *testing.T) {
	if m := NewCorsMiddleware(presetOptions); m == nil {
		t.Errorf("got %v, want %s", nil, "middleware")
		return
	}
}

func Test_isAllowedOrigin(t *testing.T) {
	c := CorsHandler{}

	c.options.AllowedOrigins = presetOptions.AllowedOrigins
	firstItem := c.options.AllowedOrigins[0]

	if resBool := c.isAllowedOrigin(firstItem); !resBool {
		t.Errorf("got %t, want %t", true, false)
	}
}

func Test_middleware(t *testing.T) {
	ctx := new(atreugo.RequestCtx)
	ctx.RequestCtx = new(fasthttp.RequestCtx)

	c := CorsHandler{}

	if err := c.middleware(ctx); err != nil && err.Error() != "CORS blocked" {
		t.Errorf("got %v, want %t", nil, err)
		return
	}

	// testing for ctx.Next()
	c.options.AllowedOrigins = []string{"*"}

	if err := c.middleware(ctx); err != nil && err.Error() != "CORS blocked" {
		t.Errorf("got %v, want %t", nil, err)
		return
	}
}

func Test_handlePreflight(t *testing.T) {
	ctx := new(atreugo.RequestCtx)
	ctx.RequestCtx = new(fasthttp.RequestCtx)

	c := CorsHandler{}
	c.options = presetOptions

	originsFirst := c.options.AllowedOrigins[0]

	ctx.Request.Header.Set("Origin", originsFirst)

	if err := c.handlePreflight(ctx); err != nil {
		t.Errorf("got %v, want %t", nil, err)
		return
	}
}
