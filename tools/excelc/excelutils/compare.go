/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

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
