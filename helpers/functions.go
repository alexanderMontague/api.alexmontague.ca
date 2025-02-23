package helpers

func Filter[T any](slice []T, f func(T) bool) []T {
	filtered := make([]T, 0)
	for _, item := range slice {
		if f(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
