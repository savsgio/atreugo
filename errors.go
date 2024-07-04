package atreugo

import (
	"fmt"
)

func wrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

func wrapErrorf(err error, message string, args ...any) error {
	message = fmt.Sprintf(message, args...)

	return wrapError(err, message)
}
