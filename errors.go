package commandregistry

import "fmt"

type UsageError struct {
	Message string
}

func (err *UsageError) Error() string {
	return err.Message
}

func UsageErrorf(format string, args ...any) error {
	return &UsageError{Message: fmt.Sprintf(format, args...)}
}
