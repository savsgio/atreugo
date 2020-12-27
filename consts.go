package atreugo

import (
	"time"
)

const (
	defaultNetwork     = "tcp4"
	defaultServerName  = "Atreugo"
	defaultReadTimeout = 20 * time.Second
)

// XRequestIDHeader header name of request id.
const XRequestIDHeader = "X-Request-ID"
