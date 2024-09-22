package fns

// NilOr
func NilOr[T any](val *T, or T) T {
	if val == nil {
		return or
	}
	return *val
}

// Ternary
func Ternary[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}
