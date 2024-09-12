package sliceutils

func Filter[T comparable](slice []T, call func(item T) bool) []T {
	res := make([]T, 0, len(slice))
	for _, item := range slice {
		if call(item) {
			res = append(res, item)
		}
	}
	return res
}

func Map[T comparable, V any](slice []T, call func(item T) V) []V {
	res := make([]V, 0, len(slice))
	for _, item := range slice {
		res = append(res, call(item))
	}
	return res
}

func Find[T comparable](slice []T, call func(item T) bool) (T, bool) {
	for _, item := range slice {
		if call(item) {
			return item, true
		}
	}
	return *new(T), false
}

func Some[T comparable](slice []T, call func(item T) bool) bool {
	for _, item := range slice {
		if call(item) {
			return true
		}
	}
	return false
}

func Every[T comparable](slice []T, call func(item T) bool) bool {
	for _, item := range slice {
		if !call(item) {
			return false
		}
	}
	return true
}

func Reduce[T comparable, V any](slice []T, call func(item T, x V) V, initValue V) V {
	res := initValue
	for _, item := range slice {
		res = call(item, res)
	}
	return res
}

func Empty[T any](in []T) bool {
	return len(in) == 0
}
