package utility

func CastToTypeOrZero[T any](v any) T {
	t, _ := v.(T)
	return t
}
