package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// UserToken is the retrieved user token from the token file.
	// It must be set using ReadToken.
	UserToken string

	// TokenDir is the directory for the token file.
	TokenDir = ".config/benchttp" // nolint:gosec // no creds
	// TokenName is the name for the token file.
	TokenName = "token.txt"

	// ErrTokenFind reports an error finding a token file.
	ErrTokenFind = errors.New("token not found")
	// ErrTokenRead reports an error reading a token file.
	ErrTokenRead = errors.New("invalid token")
	// ErrTokenSave reports an error saving a token file.
	ErrTokenSave = errors.New("cannot save token")
)

// ReadToken reads a token from a file and sets UserToken to the retrieved
// value or returns a non-nil error that is either ErrTokenFind or ErrTokenRead.
func ReadToken() error {
	// Resolve token path
	tokenPath, err := TokenPath()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTokenFind, err)
	}

	// Open token file and get its value
	b, err := os.ReadFile(tokenPath)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("%w: %s: %s", ErrTokenFind, tokenPath, err)
	default:
		return fmt.Errorf("%w: %s: %s", ErrTokenRead, tokenPath, err)
	}

	// Set UserToken to the retrieved value
	UserToken = strings.TrimSpace(string(b))
	return nil
}

// SaveToken create a file to the default path and writes the token
// into it, or returns an error that is either ErrTokenFind or ErrTokenSave.
func SaveToken(token string) error {
	// Resolve token path
	tokenPath, err := TokenPath()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTokenFind, err)
	}

	// Remove previous token file if it exists
	if _, err := os.Stat(tokenPath); !errors.Is(err, os.ErrNotExist) {
		if err := os.Remove(tokenPath); err != nil {
			return fmt.Errorf("%w: %s: %s", ErrTokenFind, tokenPath, err)
		}
	}

	// Create new token file
	f, err := os.Create(tokenPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %s", ErrTokenSave, tokenPath, err)
	}

	// Write token to file
	if _, err := f.WriteString(token + "\n"); err != nil {
		return fmt.Errorf("%w: %s: %s", ErrTokenSave, tokenPath, err)
	}

	return nil
}

// DeleteToken removes the content of the token file.
func DeleteToken() error {
	return SaveToken("")
}

// TokenPath resolves the default path for the token file. It retrieves
// the user's home directory and joins TokenDir and TokenName to it.
// If it fails it returns an ErrTokenFind error.
func TokenPath() (string, error) {
	// Retrieve user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrTokenFind, err)
	}

	// Resolve TokenDir path, making directories if needed
	dir := filepath.Join(home, TokenDir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("%w: %s", ErrTokenFind, err)
	}

	// Add TokenName to final path
	return filepath.Join(dir, TokenName), nil
}
