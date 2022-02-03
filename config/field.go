package config

type Field string

const (
	FieldMethod        Field = "method"
	FieldURL           Field = "url"
	FieldTimeout       Field = "timeout"
	FieldRequests      Field = "requests"
	FieldConcurrency   Field = "concurrency"
	FieldGlobalTimeout Field = "globalTimeout"
)

func IsField(v string) bool {
	switch Field(v) {
	case FieldMethod, FieldURL, FieldTimeout, FieldRequests,
		FieldConcurrency, FieldGlobalTimeout:
		return true
	}
	return false
}
