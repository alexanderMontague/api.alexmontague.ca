package helpers

import "time"

func Filter[T any](slice []T, f func(T) bool) []T {
	filtered := make([]T, 0)
	for _, item := range slice {
		if f(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func Map[T any, R any](ts []T, fn func(T) R) []R {
	res := make([]R, len(ts))
	for i, t := range ts {
		res[i] = fn(t)
	}
	return res
}

func Find[T any](slice []T, fn func(T) bool) *T {
	for i := range slice {
		if fn(slice[i]) {
			return &slice[i]
		}
	}
	return nil
}

func GetCurrentESTDate() string {
	now := time.Now()
	loc, _ := time.LoadLocation("America/New_York")
	return now.In(loc).Format("2006-01-02")
}
