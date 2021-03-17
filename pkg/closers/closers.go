package closers

import (
	"fmt"
	"io"
)

func SafeClose(closer io.Closer, previousError error) error {
	closeErr := closer.Close()
	if closeErr == nil {
		return previousError
	}
	if previousError == nil {
		return closeErr
	}
	return fmt.Errorf("connection closure error %v shadowed by previous error %w", closeErr, previousError)
}
