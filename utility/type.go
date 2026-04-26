package utility

func CastToTypeOrZero[T any](v any) T {
	var zero T
	if v == nil {
		return zero
	}

	t, ok := v.(T)
	if !ok {
		return zero
	}

	return t
}
