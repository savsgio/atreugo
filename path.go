package atreugo

import (
	"time"

	"github.com/valyala/fasthttp"
)

func (p *Path) Middlewares(middlewares Middlewares) *Path {
	p.middlewares = middlewares

	return p
}

func (p *Path) UseBefore(fns ...Middleware) *Path {
	p.middlewares.Before = append(p.middlewares.Before, fns...)

	return p
}

func (p *Path) UseAfter(fns ...Middleware) *Path {
	p.middlewares.After = append(p.middlewares.After, fns...)

	return p
}

func (p *Path) SkipMiddlewares(fns ...Middleware) *Path {
	p.middlewares.Skip = append(p.middlewares.Skip, fns...)

	return p
}

func (p *Path) Timeout(timeout time.Duration, msg string) *Path {
	p.withTimeout = true
	p.timeout = timeout
	p.timeoutMsg = msg
	p.timeoutCode = fasthttp.StatusRequestTimeout

	return p
}

func (p *Path) TimeoutCode(timeout time.Duration, msg string, statusCode int) *Path {
	p.withTimeout = true
	p.timeout = timeout
	p.timeoutMsg = msg
	p.timeoutCode = statusCode

	return p
}
