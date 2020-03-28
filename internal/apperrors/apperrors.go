package apperrors

import (
	"fmt"
	"strings"
)

func OfflineError(service string) error {
	return fmt.Errorf("service offline: %s", strings.ToLower(service))
}

type FatalError struct {
	Err error
}

func (e FatalError) Error() string {
	return fmt.Sprintf(`fatal error: %s`, e.Err)
}
