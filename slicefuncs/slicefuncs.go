package slicefuncs

func Map[T any, V any](in []T, transform func(T) V) []V {
	out := []V{}

	for _, v := range in {
		out = append(out, transform(v))
	}
	return out
}

func Filter[T comparable](in []T, compare func(T) bool) []T {
	out := []T{}

	for _, v := range in {
		if compare(v) {
			out = append(out, v)
		}

	}
	return out
}
