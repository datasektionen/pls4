package errors

import (
	"github.com/a-h/templ"
)

type ErrorComponent struct {
	Code int
	templ.Component
}

var _ templ.Component = ErrorComponent{}

func Error(statusCode int, messages ...string) ErrorComponent {
	return ErrorComponent{statusCode, errorComponent(statusCode, messages...)}
}
