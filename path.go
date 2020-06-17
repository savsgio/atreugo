package atreugo

import (
	"time"

	"github.com/valyala/fasthttp"
)

// Middlewares defines the middlewares (before, after and skip) in the order in which you want to execute them
// only for the view
//
// WARNING: The previous middlewares configuration could be overridden.
func (p *Path) Middlewares(middlewares Middlewares) *Path {
	p.middlewares = middlewares

	p.router.handlePath(p)

	return p
}

// UseBefore registers the middlewares in the order in which you want to execute them
// only before the execution of the view.
func (p *Path) UseBefore(fns ...Middleware) *Path {
	p.middlewares.Before = append(p.middlewares.Before, fns...)

	p.router.handlePath(p)

	return p
}

// UseAfter registers the middlewares in the order in which you want to execute them
// only after the execution of the view.
func (p *Path) UseAfter(fns ...Middleware) *Path {
	p.middlewares.After = append(p.middlewares.After, fns...)

	p.router.handlePath(p)

	return p
}

// SkipMiddlewares registers the middlewares that you want to skip only when executing the view.
func (p *Path) SkipMiddlewares(fns ...Middleware) *Path {
	p.middlewares.Skip = append(p.middlewares.Skip, fns...)

	p.router.handlePath(p)

	return p
}

// Timeout sets the timeout and the error message to the view, which returns StatusRequestTimeout
// error with the given msg to the client if view didn't return during
// the given duration.
//
// The returned view may return StatusTooManyRequests error with the given
// msg to the client if there are more concurrent views are running
// at the moment than Server.Concurrency.
func (p *Path) Timeout(timeout time.Duration, msg string) *Path {
	p.withTimeout = true
	p.timeout = timeout
	p.timeoutMsg = msg
	p.timeoutCode = fasthttp.StatusRequestTimeout

	p.router.handlePath(p)

	return p
}

// TimeoutCode sets the timeout and the error message to the view, which returns an error with
// the given msg and status code to the client if view didn't return during
// the given duration.
//
// The returned view may return StatusTooManyRequests error with the given
// msg to the client if there are more concurrent views are running
// at the moment than Server.Concurrency.
func (p *Path) TimeoutCode(timeout time.Duration, msg string, statusCode int) *Path {
	p.withTimeout = true
	p.timeout = timeout
	p.timeoutMsg = msg
	p.timeoutCode = statusCode

	p.router.handlePath(p)

	return p
}
