package argsware

const (
	ValidationErrorValueNotSet = 1<<16 + iota
	ValidationErrorValueTooSmall
	ValidationErrorValueTooBig
	ValidationErrorValueTooShort
	ValidationErrorValueTooLong
	ValidationErrorValueNotMatch
)

type (
	ArgsError struct {
		Api    string `json:"api"`
		Param  string `json:"param"`
		Reason string `json:"reason"`
	}

	ValidationError struct {
		kind  int
		field string
	}
)

func NewArgsError(api string, param string, reason string) *ArgsError {
	return &ArgsError{
		Api:    api,
		Param:  param,
		Reason: reason,
	}
}

func (e *ArgsError) Error() string {
	return "[argsWare] " + e.Api + " | " + e.Param + " | " + e.Reason
}

func NewValidationError(id int, field string) error {
	return &ValidationError{id, field}
}

func (e *ValidationError) Error() string {
	kindStr := ""
	switch e.kind {
	case ValidationErrorValueNotSet:
		kindStr = " not set"
	case ValidationErrorValueTooBig:
		kindStr = " too big"
	case ValidationErrorValueTooLong:
		kindStr = " too long"
	case ValidationErrorValueTooSmall:
		kindStr = " too small"
	case ValidationErrorValueTooShort:
		kindStr = " too short"
	case ValidationErrorValueNotMatch:
		kindStr = " not match"
	}
	return e.field + kindStr
}

func (e *ValidationError) Kind() int {
	return e.kind
}

func (e *ValidationError) Field() string {
	return e.field
}
