package atreugo

import (
	"net"
	"time"

	"github.com/fasthttp/router"
	logger "github.com/savsgio/go-logger"
	"github.com/valyala/fasthttp"
)

// private

// public

// FasthttpConfig fasthttp server configuration
//
// It is a copy from https://godoc.org/github.com/valyala/fasthttp#Server without the fields:
// - Handler: It's created internaly by atreugo
// - TCPKeepalive: Not supported yet (You can implemented with a custom Listener and pass it directly to Serve)
// - TCPKeepalivePeriod: Not supported yet (You can implemented with a custom Listener and pass it directly to Serve)
// - Logger: It's created internaly by atreugo
type FasthttpConfig struct {
	// Server name for sending in response headers.
	//
	// Default server name is used if left blank.
	Name string

	// The maximum number of concurrent connections the server may serve.
	//
	// DefaultConcurrency is used if not set.
	Concurrency int

	// Whether to disable keep-alive connections.
	//
	// The server will close all the incoming connections after sending
	// the first response to client if this option is set to true.
	//
	// By default keep-alive connections are enabled.
	DisableKeepalive bool

	// Per-connection buffer size for requests' reading.
	// This also limits the maximum header size.
	//
	// Increase this buffer if your clients send multi-KB RequestURIs
	// and/or multi-KB headers (for example, BIG cookies).
	//
	// Default buffer size is used if not set.
	ReadBufferSize int

	// Per-connection buffer size for responses' writing.
	//
	// Default buffer size is used if not set.
	WriteBufferSize int

	// Maximum duration for reading the full request (including body).
	//
	// This also limits the maximum duration for idle keep-alive
	// connections.
	//
	// By default request read timeout is unlimited if Graceful Shutdown is set to false,
	// unless will use the default ReadTimeout
	ReadTimeout time.Duration

	// Maximum duration for writing the full response (including body).
	//
	// By default response write timeout is unlimited.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alive is enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used.
	IdleTimeout time.Duration

	// Maximum number of concurrent client connections allowed per IP.
	//
	// By default unlimited number of concurrent connections
	// may be established to the server from a single IP address.
	MaxConnsPerIP int

	// Maximum number of requests served per connection.
	//
	// The server closes connection after the last request.
	// 'Connection: close' header is added to the last response.
	//
	// By default unlimited number of requests may be served per connection.
	MaxRequestsPerConn int

	// Maximum keep-alive connection lifetime.
	//
	// The server closes keep-alive connection after its' lifetime
	// expiration.
	//
	// See also ReadTimeout for limiting the duration of idle keep-alive
	// connections.
	//
	// By default keep-alive connection lifetime is unlimited.
	MaxKeepaliveDuration time.Duration

	// Maximum request body size.
	//
	// The server rejects requests with bodies exceeding this limit.
	//
	// Request body size is limited by DefaultMaxRequestBodySize by default.
	MaxRequestBodySize int

	// Aggressively reduces memory usage at the cost of higher CPU usage
	// if set to true.
	//
	// Try enabling this option only if the server consumes too much memory
	// serving mostly idle keep-alive connections. This may reduce memory
	// usage by more than 50%.
	//
	// Aggressive memory usage reduction is disabled by default.
	ReduceMemoryUsage bool

	// Rejects all non-GET requests if set to true.
	//
	// This option is useful as anti-DoS protection for servers
	// accepting only GET requests. The request size is limited
	// by ReadBufferSize if GetOnly is set.
	//
	// Server accepts all the requests by default.
	GetOnly bool

	// Logs all errors, including the most frequent
	// 'connection reset by peer', 'broken pipe' and 'connection timeout'
	// errors. Such errors are common in production serving real-world
	// clients.
	//
	// By default the most frequent errors such as
	// 'connection reset by peer', 'broken pipe' and 'connection timeout'
	// are suppressed in order to limit output log traffic.
	LogAllErrors bool

	// Header names are passed as-is without normalization
	// if this option is set.
	//
	// Disabled header names' normalization may be useful only for proxying
	// incoming requests to other servers expecting case-sensitive
	// header names. See https://github.com/valyala/fasthttp/issues/57
	// for details.
	//
	// By default request and response header names are normalized, i.e.
	// The first letter and the first letters following dashes
	// are uppercased, while all the other letters are lowercased.
	// Examples:
	//
	//     * HOST -> Host
	//     * content-type -> Content-Type
	//     * cONTENT-lenGTH -> Content-Length
	DisableHeaderNamesNormalizing bool

	// SleepWhenConcurrencyLimitsExceeded is a duration to be slept of if
	// the concurrency limit in exceeded (default [when is 0]: don't sleep
	// and accept new connections immidiatelly).
	SleepWhenConcurrencyLimitsExceeded time.Duration

	// NoDefaultServerHeader, when set to true, causes the default Server header
	// to be excluded from the Response.
	//
	// The default Server header value is the value of the Name field or an
	// internal default value in its absence. With this option set to true,
	// the only time a Server header will be sent is if a non-zero length
	// value is explicitly provided during a request.
	NoDefaultServerHeader bool

	// NoDefaultContentType, when set to true, causes the default Content-Type
	// header to be excluded from the Response.
	//
	// The default Content-Type header value is the internal default value. When
	// set to true, the Content-Type will not be present.
	NoDefaultContentType bool

	// ConnState specifies an optional callback function that is
	// called when a client connection changes state. See the
	// ConnState type and associated constants for details.
	ConnState func(net.Conn, fasthttp.ConnState)

	// KeepHijackedConns is an opt-in disable of connection
	// close by fasthttp after connections' HijackHandler returns.
	// This allows to save goroutines, e.g. when fasthttp used to upgrade
	// http connections to WS and connection goes to another handler,
	// which will close it when needed.
	KeepHijackedConns bool
}

// Config config for Atreugo
type Config struct {
	Host      string
	Port      int
	TLSEnable bool
	CertKey   string
	CertFile  string

	// Default: atreugo
	LogName string

	// See levels in https://github.com/savsgio/go-logger#levels
	LogLevel string

	// Compress transparently the response body generated by handler if the request contains 'gzip' or 'deflate'
	// in 'Accept-Encoding' header.
	Compress bool

	// Shutdown gracefully shuts down the server without interrupting any active connections.
	// Shutdown works by first closing all open listeners and then waiting indefinitely for all connections to return to idle and then shut down.
	GracefulShutdown bool

	// Custom handler to process when not matching route is found
	NotFoundHandler fasthttp.RequestHandler

	// fasthttp server configuration
	Fasthttp *FasthttpConfig
}

// Atreugo struct for make up a server
type Atreugo struct {
	lnAddr string

	server            *fasthttp.Server
	router            *router.Router
	beforeMiddlewares []Middleware
	afterMiddlewares  []Middleware
	log               *logger.Logger
	cfg               *Config
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
