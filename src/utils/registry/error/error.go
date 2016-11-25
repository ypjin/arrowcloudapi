package error

import (
	"fmt"
)

// Error : if response is returned but the status code is not 200, an Error instance will be returned
type Error struct {
	StatusCode int
	Detail     string
}

// Error returns the details as string
func (e *Error) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, e.Detail)
}
