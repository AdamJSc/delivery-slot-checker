package apperrors

import (
	"fmt"
	"strings"
)

// OfflineError represents a failed attempt to connect to a merchant
type OfflineError struct {
	Merchant string
}

func (e OfflineError) Error() string {
	return fmt.Sprintf("merchant offline: %s", strings.ToLower(e.Merchant))
}

// FatalError represents an error that should force the stoppage of the job runner
type FatalError struct {
	Err error
}

func (e FatalError) Error() string {
	return fmt.Sprintf(`fatal error: %s`, e.Err)
}
