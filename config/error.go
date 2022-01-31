package config

import (
	"fmt"
)

type ErrInvalid struct {
	invalidValues	[]error
}

func (e *ErrInvalid) Error() string {
	errorMessage := "Invalid value(s) provided:\n"
	for _, err := range e.invalidValues {
		errorMessage += err.Error() + "\n"
	}
	return fmt.Sprintf(errorMessage)
}
