package apperrors

import (
	"fmt"
	"strings"
)

// OfflineError represents a failed attempt to connect to a merchant
func OfflineError(merchant string) error {
	return fmt.Errorf("merchant offline: %s", strings.ToLower(merchant))
}

// FatalError represents an error that should force the stoppage of the job runner
type FatalError struct {
	Err error
}

func (e FatalError) Error() string {
	return fmt.Sprintf(`fatal error: %s`, e.Err)
}
