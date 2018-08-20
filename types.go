package atreugo

import (
	"github.com/erikdubbelboer/fasthttp"
	"github.com/savsgio/go-logger"
	"github.com/thehowl/fasthttprouter"
)

// private

// public

// Config config for Atreugo
type Config struct {
	Host           string
	Port           int
	LogLevel       string
	Compress       bool
	TLSEnable      bool
	CertKey        string
	CertFile       string
	GracefulEnable bool
}

// Atreugo struct for make up a server
type Atreugo struct {
	server      *fasthttp.Server
	router      *fasthttprouter.Router
	middlewares []Middleware
	log         *logger.Logger
	cfg         *Config
}

// RequestCtx context wrapper for fasthttp.RequestCtx to adds extra funtionality
type RequestCtx struct {
	*fasthttp.RequestCtx
}

// View must process incoming requests.
type View func(ctx *RequestCtx) error

// Middleware must process all incoming requests before defined views.
type Middleware func(ctx *RequestCtx) (int, error)

// JSON is a map whose key is a string and whose value an interface
type JSON map[string]interface{}
