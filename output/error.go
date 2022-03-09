package output

import (
	"errors"
	"net/http"
	"strings"

	"github.com/benchttp/runner/output/export"
)

var (
	// ErrInvalidStrategy reports an unknown strategy set.
	ErrInvalidStrategy = errors.New("invalid strategy")

	// ErrTplFailTriggered a fail triggered by a user using the function
	// {{ fail }} in an output template.
	ErrTplFailTriggered = errors.New("test failed")

	errTemplateEmpty  = errors.New("empty template")
	errTemplateSyntax = errors.New("template syntax error")
)

// ExportError is the error type returned by Report.Export.
type ExportError struct {
	Errors []error
}

// Error returns the joined errors of ExportError as a string.
func (e *ExportError) Error() string {
	const sep = "\n  - "

	var b strings.Builder
	b.WriteString("output:")
	for _, err := range e.Errors {
		b.WriteString(sep)
		b.WriteString(err.Error())
	}

	return b.String()
}

// HasAuthError returns true if any of the errors in ExportError
// is an export.HTTPResponseError with code http.StatusUnauthorized.
func (e *ExportError) HasAuthError() bool {
	for _, err := range e.Errors {
		if !errors.Is(err, export.ErrHTTPResponse) {
			continue
		}
		var errCode *export.HTTPResponseError
		if !errors.As(err, &errCode) || errCode == nil {
			continue
		}
		return errCode.Code == http.StatusUnauthorized
	}
	return false
}
