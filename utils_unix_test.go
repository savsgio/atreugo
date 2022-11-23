//go:build !windows
// +build !windows

package atreugo

import (
	"os"
	"path"
	"testing"

	"github.com/savsgio/gotils/bytes"
	"github.com/valyala/fasthttp/prefork"
)

func Test_chmodFileToSocket(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error: %v", err)
	}

	filepath := path.Join(cwd, "atreugo-test-"+string(bytes.Rand(make([]byte, 10)))+".sock")

	f, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}

	defer func() {
		f.Close()
		os.Remove(filepath)
	}()

	if err := chmodFileToSocket(filepath); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := chmodFileToSocket("243sdf$T%&$/"); err == nil {
		t.Errorf("Expected error for invalid file path")
	}
}

func Test_newPreforkServer(t *testing.T) {
	cfg := Config{
		Logger:           testLog,
		GracefulShutdown: false,
	}

	t.Run("Normal", func(t *testing.T) {
		s := New(cfg)
		sPrefork := newPreforkServer(s).(*prefork.Prefork) // nolint:forcetypeassert

		testPerforkServer(t, s, sPrefork)

		if !isEqual(sPrefork.ServeFunc, s.Serve) {
			t.Errorf("Prefork.ServeFunc == %p, want %p", sPrefork.ServeFunc, s.ServeGracefully)
		}
	})

	t.Run("Graceful", func(t *testing.T) {
		cfg.GracefulShutdown = true

		s := New(cfg)
		sPrefork := newPreforkServer(s).(*prefork.Prefork) // nolint:forcetypeassert

		if !isEqual(sPrefork.ServeFunc, s.ServeGracefully) {
			t.Errorf("Prefork.ServeFunc == %p, want %p", sPrefork.ServeFunc, s.ServeGracefully)
		}
	})
}
