package atreugo

import (
	"io"
	"net"
	"os"

	logger "github.com/savsgio/go-logger/v2"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
)

var tcpNetworks = []string{"tcp", "tcp4", "tcp6"}
var validNetworks = append(tcpNetworks, []string{"unix"}...)

// New create a new instance of Atreugo Server.
func New(cfg Config) *Atreugo {
	if cfg.Network != "" && !gotils.StringSliceInclude(validNetworks, cfg.Network) {
		panic("Invalid network: " + cfg.Network)
	}

	if cfg.Network == "" {
		cfg.Network = defaultNetwork
	}

	if cfg.Name == "" {
		cfg.Name = defaultServerName
	}

	if cfg.LogName == "" {
		cfg.LogName = defaultLogName
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = logger.INFO
	}

	if cfg.LogOutput == nil {
		cfg.LogOutput = os.Stderr
	}

	if cfg.GracefulShutdown && cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}

	cfg.socketFileMode = 0666
	log := logger.New(cfg.LogName, cfg.LogLevel, cfg.LogOutput)

	r := newRouter(log, cfg.ErrorView)

	if cfg.NotFoundView != nil {
		r.router.NotFound = viewToHandler(cfg.NotFoundView, r.errorView)
	}

	if cfg.MethodNotAllowedView != nil {
		r.router.MethodNotAllowed = viewToHandler(cfg.MethodNotAllowedView, r.errorView)
	}

	if cfg.PanicView != nil {
		r.router.PanicHandler = func(ctx *fasthttp.RequestCtx, err interface{}) {
			actx := AcquireRequestCtx(ctx)
			cfg.PanicView(actx, err)
			ReleaseRequestCtx(actx)
		}
	}

	server := &Atreugo{
		server: newFasthttpServer(cfg, log),
		log:    log,
		cfg:    cfg,
		Router: r,
	}

	return server
}

func newFasthttpServer(cfg Config, log fasthttp.Logger) *fasthttp.Server {
	return &fasthttp.Server{
		Name:                               cfg.Name,
		HeaderReceived:                     cfg.HeaderReceived,
		Concurrency:                        cfg.Concurrency,
		DisableKeepalive:                   cfg.DisableKeepalive,
		ReadBufferSize:                     cfg.ReadBufferSize,
		WriteBufferSize:                    cfg.WriteBufferSize,
		ReadTimeout:                        cfg.ReadTimeout,
		WriteTimeout:                       cfg.WriteTimeout,
		IdleTimeout:                        cfg.IdleTimeout,
		MaxConnsPerIP:                      cfg.MaxConnsPerIP,
		MaxRequestsPerConn:                 cfg.MaxRequestsPerConn,
		MaxRequestBodySize:                 cfg.MaxRequestBodySize,
		ReduceMemoryUsage:                  cfg.ReduceMemoryUsage,
		GetOnly:                            cfg.GetOnly,
		DisablePreParseMultipartForm:       cfg.DisablePreParseMultipartForm,
		LogAllErrors:                       cfg.LogAllErrors,
		DisableHeaderNamesNormalizing:      cfg.DisableHeaderNamesNormalizing,
		SleepWhenConcurrencyLimitsExceeded: cfg.SleepWhenConcurrencyLimitsExceeded,
		NoDefaultServerHeader:              cfg.NoDefaultServerHeader,
		NoDefaultDate:                      cfg.NoDefaultDate,
		NoDefaultContentType:               cfg.NoDefaultContentType,
		ConnState:                          cfg.ConnState,
		KeepHijackedConns:                  cfg.KeepHijackedConns,
		Logger:                             log,
	}
}

func (s *Atreugo) handler() fasthttp.RequestHandler {
	handler := s.router.Handler

	if len(s.virtualHosts) > 0 {
		handler = func(ctx *fasthttp.RequestCtx) {
			hostname := gotils.B2S(ctx.URI().Host())

			if h := s.virtualHosts[hostname]; h != nil {
				h(ctx)
			} else {
				s.router.Handler(ctx)
			}
		}
	}

	if s.cfg.Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	return handler
}

// SaveMatchedRoutePath if enabled, adds the matched route path onto the ctx.UserValue context
// before invoking the handler.
// The matched route path is only added to handlers of routes that were
// registered when this option was enabled.
//
// It's deactivated by default.
func (s *Atreugo) SaveMatchedRoutePath(v bool) {
	s.router.SaveMatchedRoutePath = v
}

// RedirectTrailingSlash enables/disables automatic redirection if the current route
// can't be matched but a handler for the path with (without) the trailing slash exists.
// For example if /foo/ is requested but a route only exists for /foo, the
// client is redirected to /foo with http status code 301 for GET requests
// and 307 for all other request methods.
//
// It's activated by default.
func (s *Atreugo) RedirectTrailingSlash(v bool) {
	s.router.RedirectTrailingSlash = v
}

// RedirectFixedPath if enabled, the router tries to fix the current request path, if no
// handle is registered for it.
// First superfluous path elements like ../ or // are removed.
// Afterwards the router does a case-insensitive lookup of the cleaned path.
// If a handle can be found for this route, the router makes a redirection
// to the corrected path with status code 301 for GET requests and 307 for
// all other request methods.
// For example /FOO and /..//Foo could be redirected to /foo.
// RedirectTrailingSlash is independent of this option.
//
// It's activated by default.
func (s *Atreugo) RedirectFixedPath(v bool) {
	s.router.RedirectFixedPath = v
}

// HandleMethodNotAllowed if enabled, the router checks if another method is allowed for the
// current route, if the current request can not be routed.
// If this is the case, the request is answered with 'Method Not Allowed'
// and HTTP status code 405.
// If no other Method is allowed, the request is delegated to the NotFound
// handler.
//
// It's activated by default.
func (s *Atreugo) HandleMethodNotAllowed(v bool) {
	s.router.HandleMethodNotAllowed = v
}

// HandleOPTIONS if enabled, the router automatically replies to OPTIONS requests.
// Custom OPTIONS handlers take priority over automatic replies.
//
// It's activated by default.
func (s *Atreugo) HandleOPTIONS(v bool) {
	s.handleOPTIONS = v
}

// ServeConn serves HTTP requests from the given connection.
//
// ServeConn returns nil if all requests from the c are successfully served.
// It returns non-nil error otherwise.
//
// Connection c must immediately propagate all the data passed to Write()
// to the client. Otherwise requests' processing may hang.
//
// ServeConn closes c before returning.
func (s *Atreugo) ServeConn(c net.Conn) error {
	s.server.Handler = s.handler()

	return s.server.ServeConn(c)
}

// Serve serves incoming connections from the given listener.
//
// Serve blocks until the given listener returns permanent error.
//
// If use a custom Listener, will be updated your atreugo configuration
// with the Listener address automatically.
func (s *Atreugo) Serve(ln net.Listener) error {
	defer ln.Close()

	s.cfg.Addr = ln.Addr().String()
	s.cfg.Network = ln.Addr().Network()
	s.server.Handler = s.handler()

	if gotils.StringSliceInclude(tcpNetworks, s.cfg.Network) {
		schema := "http"
		if s.cfg.TLSEnable {
			schema = "https"
		}

		s.log.Infof("Listening on: %s://%s/", schema, s.cfg.Addr)
	} else {
		s.log.Infof("Listening on (network: %s): %s ", s.cfg.Network, s.cfg.Addr)
	}

	if s.cfg.TLSEnable {
		return s.server.ServeTLS(ln, s.cfg.CertFile, s.cfg.CertKey)
	}

	return s.server.Serve(ln)
}

// SetLogOutput set log output of server.
func (s *Atreugo) SetLogOutput(output io.Writer) {
	s.log.SetOutput(output)
}

// NewVirtualHost returns a new sub-router for running more than one web site
// (such as company1.example.com and company2.example.com) on a single atreugo instance.
// Virtual hosts can be "IP-based", meaning that you have a different IP address
// for every web site, or "name-based", meaning that you have multiple names
// running on each IP address.
//
// The fact that they are running on the same atreugo instance is not apparent to the end user.
//
// If you pass multiples hostnames, all of them will have the same behaviour.
func (s *Atreugo) NewVirtualHost(hostnames ...string) *Router {
	if len(hostnames) == 0 {
		panic("At least 1 hostname is required")
	}

	if s.virtualHosts == nil {
		s.virtualHosts = make(map[string]fasthttp.RequestHandler)
	}

	vHost := newRouter(s.log, s.cfg.ErrorView)
	vHost.router.NotFound = s.router.NotFound
	vHost.router.MethodNotAllowed = s.router.MethodNotAllowed
	vHost.router.PanicHandler = s.router.PanicHandler

	for _, name := range hostnames {
		if s.virtualHosts[name] != nil {
			panicf("a router is already registered for virtual host '%s'", name)
		}

		s.virtualHosts[name] = vHost.router.Handler
	}

	return vHost
}
