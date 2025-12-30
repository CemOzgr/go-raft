package maths

import "cmp"

func Max[T cmp.Ordered](val1 T, val2 T) T {
	if val1 > val2 {
		return val1
	}

	return val2
}

func Min[T cmp.Ordered](val1 T, val2 T) T {
	if val1 < val2 {
		return val1
	}

	return val2
}
