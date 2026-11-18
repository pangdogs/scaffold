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

// Cloner is implemented by values that can return an independent copy of themselves.
type Cloner[T any] interface {
	Clone() T
}

// CloneSlice returns a clone of src.
func CloneSlice[S ~[]T, T Cloner[T]](src S) S {
	if src == nil {
		return nil
	}
	cloned := make([]T, len(src))
	for index, value := range src {
		cloned[index] = value.Clone()
	}
	return cloned
}

// CloneMap returns a clone of src.
func CloneMap[M ~map[K]V, K comparable, V Cloner[V]](src M) M {
	if src == nil {
		return nil
	}
	cloned := make(map[K]V, len(src))
	for key, value := range src {
		cloned[key] = value.Clone()
	}
	return cloned
}
