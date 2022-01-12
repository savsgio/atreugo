package atreugo

import (
	"os"
	"syscall"
)

const (
	defaultNetwork    = "tcp4"
	defaultServerName = "Atreugo"
)

var defaultGracefulShutdownSignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
}

// XRequestIDHeader header name of request id.
const XRequestIDHeader = "X-Request-ID"
