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

package goscr

// EntityThis 脚本化实体，支持获取This指针，用于函数模式绑定脚本函数
type EntityThis[T any] struct {
	Entity
}

func (e *EntityThis[T]) This() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityEnableUpdateThis 脚本化实体，支持Update，支持获取This指针，用于函数模式绑定脚本函数
type EntityEnableUpdateThis[T any] struct {
	EntityEnableUpdate
}

func (e *EntityEnableUpdateThis[T]) This() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityEnableLateUpdateThis 脚本化实体，支持LateUpdate，支持获取This指针，用于函数模式绑定脚本函数
type EntityEnableLateUpdateThis[T any] struct {
	EntityEnableLateUpdate
}

func (e *EntityEnableLateUpdateThis[T]) This() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityEnableUpdateAndLateUpdateThis 脚本化实体，支持Update、LateUpdate，支持获取This指针，用于函数模式绑定脚本函数
type EntityEnableUpdateAndLateUpdateThis[T any] struct {
	EntityEnableUpdateAndLateUpdate
}

func (e *EntityEnableUpdateAndLateUpdateThis[T]) This() *T {
	return e.GetReflected().Interface().(*T)
}
