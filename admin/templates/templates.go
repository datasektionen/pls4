package templates

import (
	"github.com/a-h/templ"
)

//go:generate templ generate

type ErrorComponent struct {
	Code int
	templ.Component
}

var _ templ.Component = ErrorComponent{}

func Error(statusCode int, messages ...string) ErrorComponent {
	return ErrorComponent{statusCode, errorComponent(statusCode, messages...)}
}

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func ternary[T any](condition bool, then T, elze T) T {
	if condition {
		return then
	} else {
		return elze
	}
}

