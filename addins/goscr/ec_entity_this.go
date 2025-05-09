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

func (c *EntityThis[T]) This() func() *T {
	return c.thisFunc
}

func (e *EntityThis[T]) thisFunc() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityEnableUpdateThis 脚本化实体，支持帧更新（Update），支持获取This指针，用于函数模式绑定脚本函数
type EntityEnableUpdateThis[T any] struct {
	EntityEnableUpdate
}

func (c *EntityEnableUpdateThis[T]) This() func() *T {
	return c.thisFunc
}

func (e *EntityEnableUpdateThis[T]) thisFunc() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityEnableLateUpdateThis 脚本化实体，支持帧迟滞更新（Late Update），支持获取This指针，用于函数模式绑定脚本函数
type EntityEnableLateUpdateThis[T any] struct {
	EntityEnableLateUpdate
}

func (c *EntityEnableLateUpdateThis[T]) This() func() *T {
	return c.thisFunc
}

func (e *EntityEnableLateUpdateThis[T]) thisFunc() *T {
	return e.GetReflected().Interface().(*T)
}

// EntityEnableUpdateAndLateUpdateThis 脚本化实体，支持帧更新（Update）、帧迟滞更新（Late Update），支持获取This指针，用于函数模式绑定脚本函数
type EntityEnableUpdateAndLateUpdateThis[T any] struct {
	EntityEnableUpdateAndLateUpdate
}

func (c *EntityEnableUpdateAndLateUpdateThis[T]) This() func() *T {
	return c.thisFunc
}

func (e *EntityEnableUpdateAndLateUpdateThis[T]) thisFunc() *T {
	return e.GetReflected().Interface().(*T)
}
