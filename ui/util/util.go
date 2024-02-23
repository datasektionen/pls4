package util

func Plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func Ternary[T any](condition bool, then T, elze T) T {
	if condition {
		return then
	} else {
		return elze
	}
}
