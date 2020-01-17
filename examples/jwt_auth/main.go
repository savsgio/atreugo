package main

import (
	"fmt"

	"github.com/savsgio/atreugo/v10"
	"github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
)

func init() { //nolint:gochecknoinits
	logger.SetLevel(logger.DEBUG)
}

func main() {
	config := &atreugo.Config{
		Addr: "0.0.0.0:8000",
	}
	server := atreugo.New(config)

	// Register authentication middleware at first of all
	server.UseBefore(authMiddleware)

	// Register index route
	server.Path("GET", "/", func(ctx *atreugo.RequestCtx) error {
		return ctx.HTTPResponse(fmt.Sprintf(`<h1>You are login with JWT</h1>
				JWT cookie value: %s`, ctx.Request.Header.Cookie("atreugo_jwt")))
	})

	// Register login route
	server.Path("GET", "/login", func(ctx *atreugo.RequestCtx) error {
		qUser := []byte("savsgio")
		qPasswd := []byte("mypasswd")

		jwtCookie := ctx.Request.Header.Cookie("atreugo_jwt")

		if len(jwtCookie) == 0 {
			tokenString, expireAt := generateToken(qUser, qPasswd)

			// Set cookie for domain
			cookie := fasthttp.AcquireCookie()
			defer fasthttp.ReleaseCookie(cookie)

			cookie.SetKey("atreugo_jwt")
			cookie.SetValue(tokenString)
			cookie.SetExpire(expireAt)
			ctx.Response.Header.SetCookie(cookie)
		}

		return ctx.RedirectResponse("/", ctx.Response.StatusCode())
	})

	// Run
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
