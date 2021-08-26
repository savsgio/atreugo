//go:build !windows
// +build !windows

package atreugo

import (
	"os"
	"path"
	"testing"

	"github.com/savsgio/gotils/bytes"
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
