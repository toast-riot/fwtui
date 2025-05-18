package listext

func Singleton[T any](value T) []T {
	return []T{value}
}
