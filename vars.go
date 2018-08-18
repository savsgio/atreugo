package atreugo

import "sync"

var allowedHTTPMethods = []string{"GET", "HEAD", "OPTIONS", "POST", "PUT", "PATCH", "DELETE"}

var requestCtxPool = sync.Pool{
	New: func() interface{} {
		return new(RequestCtx)
	},
}
