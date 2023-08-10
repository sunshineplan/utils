package slice

// Deduplicate removes duplicate items in slice.
func Deduplicate[S ~[]E, E comparable](s S) S {
	if s == nil {
		return nil
	}

	res := S{}
	m := make(map[E]struct{})
	for _, i := range s {
		if _, ok := m[i]; !ok {
			m[i] = struct{}{}
			res = append(res, i)
		}
	}
	return res
}
