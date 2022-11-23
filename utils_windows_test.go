//go:build windows
// +build windows

package atreugo

import (
	"testing"

	"github.com/valyala/fasthttp/prefork"
)

func Test_chmodFileToSocket(t *testing.T) {
	if err := chmodFileToSocket("filepath"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func Test_newPreforkServer(t *testing.T) {
	cfg := Config{
		Logger:           testLog,
		GracefulShutdown: false,
	}

	s := New(cfg)
	sPrefork := newPreforkServer(s).(*prefork.Prefork) // nolint:forcetypeassert

	testPerforkServer(t, s, sPrefork)

	if !isEqual(sPrefork.ServeFunc, s.Serve) {
		t.Errorf("Prefork.ServeFunc == %p, want %p", sPrefork.ServeFunc, s.Serve)
	}
}
