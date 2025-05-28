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

// EntityStateThis 脚本化实体状态This指针，用于函数模式绑定脚本函数
type EntityStateThis[T any] struct {
	EntityState
}

func (e *EntityStateThis[T]) This() func() *T {
	return e.thisFunc
}

func (e *EntityStateThis[T]) thisFunc() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityStateEnableUpdateThis 脚本化实体状态This指针，支持帧更新（Update），用于函数模式绑定脚本函数
type EntityStateEnableUpdateThis[T any] struct {
	EntityStateEnableUpdate
}

func (e *EntityStateEnableUpdateThis[T]) This() func() *T {
	return e.thisFunc
}

func (e *EntityStateEnableUpdateThis[T]) thisFunc() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityStateEnableLateUpdateThis 脚本化实体状态This指针，支持帧迟滞更新（Late Update），用于函数模式绑定脚本函数
type EntityStateEnableLateUpdateThis[T any] struct {
	EntityStateEnableLateUpdate
}

func (e *EntityStateEnableLateUpdateThis[T]) This() func() *T {
	return e.thisFunc
}

func (e *EntityStateEnableLateUpdateThis[T]) thisFunc() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityStateEnableUpdateAndLateUpdateThis 脚本化实体状态This指针，支持帧更新（Update）、帧迟滞更新（Late Update），用于函数模式绑定脚本函数
type EntityStateEnableUpdateAndLateUpdateThis[T any] struct {
	EntityStateEnableUpdateAndLateUpdate
}

func (e *EntityStateEnableUpdateAndLateUpdateThis[T]) This() func() *T {
	return e.thisFunc
}

func (e *EntityStateEnableUpdateAndLateUpdateThis[T]) thisFunc() *T {
	return e.GetReflected().Interface().(*T)
}
