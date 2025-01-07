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

// ComponentThis 脚本化组件，支持获取This指针，用于函数模式绑定脚本函数
type ComponentThis[T any] struct {
	Component
}

func (c *ComponentThis[T]) This() *T {
	return c.GetReflected().Interface().(*T)
}

// ComponentEnableUpdateThis 脚本化组件，支持帧更新（Update），支持获取This指针，用于函数模式绑定脚本函数
type ComponentEnableUpdateThis[T any] struct {
	ComponentEnableUpdate
}

func (c *ComponentEnableUpdateThis[T]) This() *T {
	return c.GetReflected().Interface().(*T)
}

// ComponentEnableLateUpdateThis 脚本化组件，支持帧迟滞更新（Late Update），支持获取This指针，用于函数模式绑定脚本函数
type ComponentEnableLateUpdateThis[T any] struct {
	ComponentEnableLateUpdate
}

func (c *ComponentEnableLateUpdateThis[T]) This() *T {
	return c.GetReflected().Interface().(*T)
}

// ComponentEnableUpdateAndLateUpdateThis 脚本化组件，支持帧更新（Update）、帧迟滞更新（Late Update），支持获取This指针，用于函数模式绑定脚本函数
type ComponentEnableUpdateAndLateUpdateThis[T any] struct {
	ComponentEnableUpdateAndLateUpdate
}

func (c *ComponentEnableUpdateAndLateUpdateThis[T]) This() *T {
	return c.GetReflected().Interface().(*T)
}
