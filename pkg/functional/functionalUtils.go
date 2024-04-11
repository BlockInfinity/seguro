package functional

// https://stackoverflow.com/a/71624929
func Map[T, U any](ts []T, f func(T) U) []U {
	result := make([]U, len(ts))
	for i := range ts {
		result[i] = f(ts[i])
	}

	return result
}

func MapWithIndex[T, U any](ts []T, f func(T, int) U) []U {
	result := make([]U, len(ts))
	for i := range ts {
		result[i] = f(ts[i], i)
	}

	return result
}

func MapWithError[T, U any](ts []T, f func(T) (U, error)) ([]U, error) {
	result := make([]U, len(ts))
	for i := range ts {
		var err error
		result[i], err = f(ts[i])
		if err != nil {
			return make([]U, 0), err
		}
	}

	return result, nil
}

// https://stackoverflow.com/a/37563128
func Filter[T any](ss []T, test func(T) bool) []T {
	ret := make([]T, 0)
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}

	return ret
}

func ArrayIncludes[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
