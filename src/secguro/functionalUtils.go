package main

// https://stackoverflow.com/a/71624929
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func MapWithError[T, U any](ts []T, f func(T) (U, error)) ([]U, error) {
	us := make([]U, len(ts))
	for i := range ts {
		var err error
		us[i], err = f(ts[i])
		if err != nil {
			return make([]U, 0), err
		}
	}
	return us, nil
}

// https://stackoverflow.com/a/37563128
func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func arrayIncludes[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
