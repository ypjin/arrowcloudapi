package registry

import (
	"net/http"
)

// Modifier modifies request
type Modifier interface {
	Modify(*http.Request) error
}
