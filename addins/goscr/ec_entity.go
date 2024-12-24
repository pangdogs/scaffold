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
	"fmt"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/ec/pt"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework"
	"github.com/elliotchance/pie/v2"
	"maps"
	"reflect"
	"strings"
)

// Entity 脚本化实体
type Entity struct {
	framework.EntityBehavior
}

// Callee 被调函数
func (c *Entity) Callee(method string) reflect.Value {
	return reflect.ValueOf(c.bindMethod(method))
}

// Awake 生命周期Awake
func (c *Entity) Awake() {
	method, _ := c.bindMethod("Awake").(func())
	if method != nil {
		method()
	}
}

// Start 生命周期Start
func (c *Entity) Start() {
	method, _ := c.bindMethod("Start").(func())
	if method != nil {
		method()
	}
}

// Shut 生命周期Shut
func (c *Entity) Shut() {
	method, _ := c.bindMethod("Shut").(func())
	if method != nil {
		method()
	}
}

// Dispose 生命周期Dispose
func (c *Entity) Dispose() {
	method, _ := c.bindMethod("Dispose").(func())
	if method != nil {
		method()
	}
}

// EntityEnableUpdate 脚本化实体，支持Update
type EntityEnableUpdate struct {
	Entity
}

// Update 生命周期Update
func (c *EntityEnableUpdate) Update() {
	method, _ := c.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

func (c *Entity) bindMethod(method string) any {
	scriptPkg, ok := c.GetPT().Extra().Get("script_pkg")
	if !ok {
		return nil
	}

	scriptIdent, ok := c.GetPT().Extra().Get("script_ident")
	if !ok {
		return nil
	}

	thisMethod := Using(c.GetService()).Solution().BindMethod(c.GetReflected().Interface(), scriptPkg.(string), scriptIdent.(string), method)
	if !thisMethod.IsValid() {
		return nil
	}

	return thisMethod.Interface()
}

// EntityEnableLateUpdate 脚本化实体，支持LateUpdate
type EntityEnableLateUpdate struct {
	Entity
}

// LateUpdate 生命周期LateUpdate
func (c *EntityEnableLateUpdate) LateUpdate() {
	method, _ := c.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}

// EntityEnableUpdateAndLateUpdate 脚本化实体，支持Update、LateUpdate
type EntityEnableUpdateAndLateUpdate struct {
	Entity
}

// Update 生命周期Update
func (c *EntityEnableUpdateAndLateUpdate) Update() {
	method, _ := c.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

// LateUpdate 生命周期LateUpdate
func (c *EntityEnableUpdateAndLateUpdate) LateUpdate() {
	method, _ := c.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}

// EntityWith 创建脚本化实体原型属性，用于注册实体原型时自定义相关属性
func EntityWith(prototype, script string, scope *ec.Scope, componentAwakeOnFirstTouch, componentUniqueID *bool, extra ...map[string]any) pt.EntityAttribute {
	return EntityWithT[Entity](prototype, script, scope, componentAwakeOnFirstTouch, componentUniqueID, extra...)
}

// EntityWithT 创建脚本化实体原型属性，用于注册实体原型时自定义相关属性
func EntityWithT[T any](prototype, script string, scope *ec.Scope, componentAwakeOnFirstTouch, componentUniqueID *bool, extra ...map[string]any) pt.EntityAttribute {
	idx := strings.LastIndexByte(script, '.')
	if idx < 0 {
		panic(fmt.Errorf("incorrect script %q format", script))
	}

	scriptPkg := script[:idx]
	scriptIdent := script[idx+1:]

	_extra := maps.Clone(pie.First(extra))
	if _extra == nil {
		_extra = map[string]any{}
	}
	_extra["script_pkg"] = scriptPkg
	_extra["script_ident"] = scriptIdent

	return pt.EntityWith(prototype, types.ZeroT[T](), scope, componentAwakeOnFirstTouch, componentUniqueID, _extra)
}
