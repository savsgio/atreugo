// +build windows

package atreugo

import (
	"os"
	"path"
	"testing"

	"github.com/savsgio/gotils/bytes"
)

func Test_chmodFileToSocket(t *testing.T) {
	if err := chmodFileToSocket("filepath"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
