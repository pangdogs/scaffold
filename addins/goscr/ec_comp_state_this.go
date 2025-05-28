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

// ComponentStateThis 脚本化组件状态This指针，用于函数模式绑定脚本函数
type ComponentStateThis[T any] struct {
	ComponentState
}

func (c *ComponentStateThis[T]) This() func() *T {
	return c.thisFunc
}

func (c *ComponentStateThis[T]) thisFunc() *T {
	return c.GetReflected().Interface().(*T)
}

// ComponentStateEnableUpdateThis 脚本化组件状态This指针，支持帧更新（Update），用于函数模式绑定脚本函数
type ComponentStateEnableUpdateThis[T any] struct {
	ComponentStateEnableUpdate
}

func (c *ComponentStateEnableUpdateThis[T]) This() func() *T {
	return c.thisFunc
}

func (c *ComponentStateEnableUpdateThis[T]) thisFunc() *T {
	return c.GetReflected().Interface().(*T)
}

// ComponentStateEnableLateUpdateThis 脚本化组件状态This指针，支持帧迟滞更新（Late Update），用于函数模式绑定脚本函数
type ComponentStateEnableLateUpdateThis[T any] struct {
	ComponentStateEnableLateUpdate
}

func (c *ComponentStateEnableLateUpdateThis[T]) This() func() *T {
	return c.thisFunc
}

func (c *ComponentStateEnableLateUpdateThis[T]) thisFunc() *T {
	return c.GetReflected().Interface().(*T)
}

// ComponentStateEnableUpdateAndLateUpdateThis 脚本化组件状态This指针，支持帧更新（Update）、帧迟滞更新（Late Update），用于函数模式绑定脚本函数
type ComponentStateEnableUpdateAndLateUpdateThis[T any] struct {
	ComponentStateEnableUpdateAndLateUpdate
}

func (c *ComponentStateEnableUpdateAndLateUpdateThis[T]) This() func() *T {
	return c.thisFunc
}

func (c *ComponentStateEnableUpdateAndLateUpdateThis[T]) thisFunc() *T {
	return c.GetReflected().Interface().(*T)
}
