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
	case FieldMethod:
	case FieldURL:
	case FieldTimeout:
	case FieldRequests:
	case FieldConcurrency:
	case FieldGlobalTimeout:
		return true
	}
	return false
}
