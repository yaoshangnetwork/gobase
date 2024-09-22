package sliceutils

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

func Filter[T any](slice []T, call func(item T) bool) []T {
	res := make([]T, 0, len(slice))
	for _, item := range slice {
		if call(item) {
			res = append(res, item)
		}
	}
	return res
}

func Map[T any, V any](slice []T, call func(item T) V) []V {
	res := make([]V, 0, len(slice))
	for _, item := range slice {
		res = append(res, call(item))
	}
	return res
}

func Find[T any](slice []T, call func(item T) bool) (T, bool) {
	for _, item := range slice {
		if call(item) {
			return item, true
		}
	}
	return *new(T), false
}

func FindIndex[T any](slice []T, call func(item T) bool) int {
	for i, item := range slice {
		if call(item) {
			return i
		}
	}
	return -1
}

func FindLastIndex[T any](slice []T, call func(item T) bool) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if call(slice[i]) {
			return i
		}
	}
	return -1
}

func Some[T any](slice []T, call func(item T) bool) bool {
	for _, item := range slice {
		if call(item) {
			return true
		}
	}
	return false
}

func Every[T any](slice []T, call func(item T) bool) bool {
	for _, item := range slice {
		if !call(item) {
			return false
		}
	}
	return true
}

func Reduce[T any, V any](slice []T, call func(item T, x V) V, initValue V) V {
	res := initValue
	for _, item := range slice {
		res = call(item, res)
	}
	return res
}

func Contains[T comparable](slice []T, val T) bool {
	return FindIndex(slice, func(item T) bool { return item == val }) != -1
}

func Empty[T any](in []T) bool {
	return len(in) == 0
}
