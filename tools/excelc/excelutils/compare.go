package excelutils

import (
	"cmp"
)

func ListEqual[T any](a, b []T, fun func(a, b T) bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !fun(a[i], b[i]) {
			return false
		}
	}

	return true
}

func MapEqual[K cmp.Ordered, V any](a, b map[K]V, fun func(a, b V) bool) bool {
	if len(a) != len(b) {
		return false
	}

	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if !fun(av, bv) {
			return false
		}
	}

	return true
}
