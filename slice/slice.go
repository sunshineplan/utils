package slice

// Reverse reverses a slice's order.
func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

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
