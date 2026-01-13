package slices

func RemoveByValue[T comparable](s []T, val T) []T {
	for i, v := range s {
		if v == val {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
