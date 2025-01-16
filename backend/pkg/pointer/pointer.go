package pointer

// Of is a helper routine that allocates a new any value
// to store v and returns a pointer to it.
func Of[T any](v T) *T {
	return &v
}

func Slice[T any](slice []T) []*T {
	res := make([]*T, len(slice))
	for i := range slice {
		res[i] = Of(slice[i])
	}
	return res
}

func PAny[T any](v *T) T {
	if v == nil {
		v = new(T)
	}

	return *v
}

func EmptySlice[T any]() []T {
	return make([]T, 0)
}

func IsNil[T any](v *T) bool {
	return v == nil
}
