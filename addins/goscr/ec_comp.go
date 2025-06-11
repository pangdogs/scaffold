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

import (
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/ec/pt"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
	"strings"
)

// ComponentScriptBehavior 脚本化组件行为
type ComponentScriptBehavior struct {
	ComponentStateEnableUpdateAndLateUpdateThis[ComponentScriptBehavior]
}

// ComponentScript 创建脚本化组件原型属性，用于注册实体原型时自定义相关属性
func ComponentScript(script string) *pt.ComponentAttribute {
	return ComponentScriptT[ComponentScriptBehavior](script)
}

// ComponentScriptT 创建脚本化组件原型属性，自定义组件状态类型，用于注册实体原型时自定义相关属性
func ComponentScriptT[T any](script string) *pt.ComponentAttribute {
	if script == "" {
		exception.Panicf("goscr: %w: script is empty", exception.ErrArgs)
	}

	idx := strings.LastIndexByte(script, '.')
	if idx < 0 {
		exception.Panicf("goscr: incorrect script %q format", script)
	}

	scriptPkg := script[:idx]
	scriptIdent := script[idx+1:]

	return pt.NewComponentAttribute(types.ZeroT[T]()).SetName(scriptIdent).SetExtra(map[string]any{"script_pkg": scriptPkg, "script_ident": scriptIdent})
}

// GetComponentScript 获取组件脚本
func GetComponentScript(entity ec.Entity, name string) func() *ComponentScriptBehavior {
	return GetComponentScriptT[*ComponentScriptBehavior](entity, name)
}

// GetComponentScriptT 获取自定义组件状态类型的脚本
func GetComponentScriptT[T interface{ This() func() T }](entity ec.Entity, name string) func() T {
	if entity == nil {
		exception.Panicf("goscr: %s: entity is nil", exception.ErrArgs)
	}

	comp := entity.GetComponent(name)
	if comp == nil {
		return nil
	}

	behavior, ok := comp.(T)
	if !ok {
		return nil
	}

	return behavior.This()
}
