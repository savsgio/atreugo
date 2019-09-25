package main

import (
	"errors"

	"github.com/savsgio/atreugo/v9"
	"github.com/valyala/fasthttp"
)

// checkTokenMiddleware middleware to check jwt token authorization
func authMiddleware(ctx *atreugo.RequestCtx) error {
	// Avoid middleware when you are going to login view
	if string(ctx.Path()) == "/login" {
		return ctx.Next()
	}

	jwtCookie := ctx.Request.Header.Cookie("atreugo_jwt")

	if len(jwtCookie) == 0 {
		return ctx.ErrorResponse(errors.New("login required"), fasthttp.StatusForbidden)
	}

	token, _, err := validateToken(string(jwtCookie))
	if err != nil {
		return err
	}

	if !token.Valid {
		return ctx.ErrorResponse(errors.New("your session is expired, login again please"), fasthttp.StatusForbidden)
	}

	return ctx.Next()
}
