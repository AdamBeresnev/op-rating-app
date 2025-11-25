package utils

import "strings"

func Ptr[T any](v T) *T {
	return &v
}

func OrZero[T comparable](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

// Returns nil on an empty or all whitespace string
func StringOrNil(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
