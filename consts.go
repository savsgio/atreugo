package atreugo

import (
	"time"
)

const defaultNetwork = "tcp4"

const defaultLogName = "atreugo"
const defaultServerName = "Atreugo"
const defaultReadTimeout = 20 * time.Second

const attachedCtxKey = "__attachedCtx__"

// XRequestIDHeader header name of request id
const XRequestIDHeader = "X-Request-ID"
