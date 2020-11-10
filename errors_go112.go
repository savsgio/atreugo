// +build go1.12

package atreugo

import (
	"fmt"
)

func wrapError(err error, message string, args ...interface{}) error {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	return fmt.Errorf("%s: %v", message, err) // nolint:errorlint
}

func errorIs(err, target error) bool {
	return err == target // nolint:errorlint
}
