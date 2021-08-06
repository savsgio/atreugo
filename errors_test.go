package atreugo

import (
	"errors"
	"testing"
)

func Test_wrapError(t *testing.T) {
	err := errors.New("some error")

	err2 := wrapError(err, "test")
	if err2 == nil {
		t.Error("Expected error, got nil")
	}

	wantStrEr := "test: some error"
	if strErr := err2.Error(); strErr != wantStrEr {
		t.Errorf("result %v, want %v", strErr, wantStrEr)
	}

	if !errors.Is(err2, err) {
		t.Error("Expected different error")
	}
}

func Test_wrapErrorf(t *testing.T) {
	err := errors.New("some error")

	err2 := wrapErrorf(err, "test %s", "fail")
	if err2 == nil {
		t.Error("Expected error, got nil")
	}

	wantStrEr := "test fail: some error"
	if strErr := err2.Error(); strErr != wantStrEr {
		t.Errorf("result %v, want %v", strErr, wantStrEr)
	}

	if !errors.Is(err2, err) {
		t.Error("Expected different error")
	}
}
