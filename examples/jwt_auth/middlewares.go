package main

import (
	"errors"

	"github.com/savsgio/atreugo/v8"
	"github.com/valyala/fasthttp"
)

// checkTokenMiddleware middleware to check jwt token authorization
func authMiddleware(ctx *atreugo.RequestCtx) (int, error) {
	// Avoid middleware when you are going to login view
	if string(ctx.Path()) == "/login" {
		return fasthttp.StatusOK, nil
	}

	jwtCookie := ctx.Request.Header.Cookie("atreugo_jwt")

	if len(jwtCookie) == 0 {
		return fasthttp.StatusForbidden, errors.New("login required")
	}

	token, _, err := validateToken(string(jwtCookie))

	if !token.Valid {
		return fasthttp.StatusForbidden, errors.New("your session is expired, login again please")
	}

	return fasthttp.StatusOK, err
}
