package provider

func toPtr[T any](v T) *T {
	return &v
}

func each[T any, V any](ss []T, fn func(T) V) []V {
	var vs = make([]V, len(ss))
	for i := 0; i < len(ss); i++ {
		vs[i] = fn(ss[i])
	}

	return vs
}
