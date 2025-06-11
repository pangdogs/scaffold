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

// EntityScriptBehavior 脚本化实体行为
type EntityScriptBehavior struct {
	EntityStateEnableUpdateAndLateUpdateThis[EntityScriptBehavior]
}

// EntityScript 创建脚本化实体原型属性，用于注册实体原型时自定义相关属性
func EntityScript(prototype, script string) *pt.EntityAttribute {
	return EntityScriptT[EntityScriptBehavior](prototype, script)
}

// EntityScriptT 创建脚本化实体原型属性，自定义实体状态类型，用于注册实体原型时自定义相关属性
func EntityScriptT[T any](prototype, script string) *pt.EntityAttribute {
	if prototype == "" {
		exception.Panicf("goscr: %w: prototype is empty", exception.ErrArgs)
	}

	if script == "" {
		exception.Panicf("goscr: %w: script is empty", exception.ErrArgs)
	}

	idx := strings.LastIndexByte(script, '.')
	if idx < 0 {
		exception.Panicf("goscr: incorrect script %q format", script)
	}

	scriptPkg := script[:idx]
	scriptIdent := script[idx+1:]

	return pt.NewEntityAttribute(prototype).SetInstance(types.ZeroT[T]()).SetExtra(map[string]any{"script_pkg": scriptPkg, "script_ident": scriptIdent})
}

// GetEntityScript 获取实体脚本
func GetEntityScript(entity ec.Entity) func() *EntityScriptBehavior {
	return GetEntityScriptT[*EntityScriptBehavior](entity)
}

// GetEntityScriptT 获取自定义实体状态类型的脚本
func GetEntityScriptT[T interface{ This() func() T }](entity ec.Entity) func() T {
	if entity == nil {
		exception.Panicf("goscr: %s: entity is nil", exception.ErrArgs)
	}

	behavior, ok := entity.(T)
	if !ok {
		return nil
	}

	return behavior.This()
}
