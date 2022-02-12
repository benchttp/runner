package config

const (
	FieldMethod         = "method"
	FieldURL            = "url"
	FieldHeader         = "header"
	FieldRequests       = "requests"
	FieldConcurrency    = "concurrency"
	FieldInterval       = "interval"
	FieldRequestTimeout = "requestTimeout"
	FieldGlobalTimeout  = "globalTimeout"
)

func IsField(v string) bool {
	switch v {
	case FieldMethod, FieldURL, FieldHeader, FieldRequests,
		FieldConcurrency, FieldInterval, FieldRequestTimeout,
		FieldGlobalTimeout:
		return true
	}
	return false
}
