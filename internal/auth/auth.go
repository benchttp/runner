package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	// ErrTokenFind reports an error finding a token file.
	ErrTokenFind = errors.New("token not found")
	// ErrTokenRead reports an error reading a token file.
	ErrTokenRead = errors.New("invalid token")
	// ErrTokenSave reports an error saving a token file.
	ErrTokenSave = errors.New("cannot save token")
)

// ReadToken reads a token from a file or returns a non-nil error
// that is either ErrTokenFind or ErrTokenRead.
func ReadToken(path string) (string, error) {
	b, err := os.ReadFile(path)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return "", fmt.Errorf("%w: %s: %s", ErrTokenFind, path, err)
	default:
		return "", fmt.Errorf("%w: %s: %s", ErrTokenRead, path, err)
	}

	return strings.TrimSpace(string(b)), nil
}

// SaveToken create a file th given path and writes the token to it.
func SaveToken(path, token string) error {
	// Remove previous token file if exists
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("%w: %s: %s", ErrTokenSave, path, err)
		}
	}

	// Create new token file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("%w: %s: %s", ErrTokenSave, path, err)
	}

	// Write token to file
	if _, err := f.WriteString(token + "\n"); err != nil {
		return fmt.Errorf("%w: %s: %s", ErrTokenSave, path, err)
	}

	return nil
}
