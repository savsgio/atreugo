// +build !windows

package atreugo

import (
	"os"
)

func chmodFileToSocket(filepath string) error {
	if err := os.Chmod(filepath, os.ModeSocket); err != nil {
		return wrapErrorf(err, "cannot chmod %#o for %q", os.ModeSocket, filepath)
	}

	return nil
}
