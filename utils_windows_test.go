//go:build windows
// +build windows

package atreugo

import "testing"

func Test_chmodFileToSocket(t *testing.T) {
	if err := chmodFileToSocket("filepath"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
