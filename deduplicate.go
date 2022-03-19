package utils

// Deduplicate removes duplicate items in slice
func Deduplicate[T comparable](s []T) []T {
	res := []T{}
	m := make(map[T]struct{})
	for _, i := range s {
		if _, ok := m[i]; !ok {
			m[i] = struct{}{}
			res = append(res, i)
		}
	}
	return res
}
