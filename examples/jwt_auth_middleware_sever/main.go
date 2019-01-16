package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/savsgio/atreugo/v7"
	"github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
)

var jwtSignKey = []byte("TestForFasthttpWithJWT")

type userCredential struct {
	Username []byte `json:"username"`
	Password []byte `json:"password"`
	jwt.StandardClaims
}

func init() {
	logger.SetLevel(logger.DEBUG)
}

func generateToken(username []byte, password []byte) (string, time.Time) {
	logger.Debugf("Create new token for user %s", username)

	expireAt := time.Now().Add(1 * time.Minute)

	// Embed User information to `token`
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS512, &userCredential{
		Username: username,
		Password: password,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireAt.Unix(),
		},
	})

	// token -> string. Only server knows the secret.
	tokenString, err := newToken.SignedString(jwtSignKey)
	if err != nil {
		logger.Error(err)
	}

	return tokenString, expireAt
}

func validateToken(requestToken string) (*jwt.Token, *userCredential, error) {
	logger.Debug("Validating token...")

	user := &userCredential{}
	token, err := jwt.ParseWithClaims(requestToken, user, func(token *jwt.Token) (interface{}, error) {
		return jwtSignKey, nil
	})

	return token, user, err
}

// checkTokenMiddleware middleware to check jwt token authorization
func checkTokenMiddleware(ctx *atreugo.RequestCtx) (int, error) {
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

func main() {
	config := &atreugo.Config{
		Host: "0.0.0.0",
		Port: 8000,
	}
	server := atreugo.New(config)

	server.UseMiddleware(checkTokenMiddleware)

	server.Path("GET", "/", func(ctx *atreugo.RequestCtx) error {
		return ctx.HTTPResponse(fmt.Sprintf(`<h1>You are login with JWT</h1>
				JWT cookie value: %s`, ctx.Request.Header.Cookie("atreugo_jwt")))
	})

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

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
