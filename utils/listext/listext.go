package listext

func Singleton[T any](value T) []T {
	return []T{value}
}

func Slice[T any](values ...T) []T {
	return values
}
