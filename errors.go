package atreugo

import (
	"errors"
	"fmt"
)

func wrapError(err error, message string, args ...interface{}) error {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	return fmt.Errorf("%s: %w", message, err)
}

func errorIs(err, target error) bool {
	return errors.Is(err, target)
}
