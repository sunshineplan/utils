package slice

// Deduplicate removes duplicate elements while preserving order.
func Deduplicate[S ~[]E, E comparable](s S) S {
	if len(s) == 0 {
		return nil
	}
	m := make(map[E]struct{}, len(s))
	res := make(S, 0, len(s))
	for _, i := range s {
		if _, ok := m[i]; !ok {
			m[i] = struct{}{}
			res = append(res, i)
		}
	}
	return res
}
