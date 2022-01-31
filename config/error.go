package config

type ErrInvalid struct {
	invalidValues []error
}

func (e *ErrInvalid) Error() string {
	errorMessage := "Invalid value(s) provided:\n"
	for _, err := range e.invalidValues {
		errorMessage += err.Error() + "\n"
	}
	return errorMessage
}
