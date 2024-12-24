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
	"git.golaxy.org/core/ec/pt"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework"
	"github.com/elliotchance/pie/v2"
	"maps"
	"reflect"
	"strings"
)

// Component 脚本化组件
type Component struct {
	framework.ComponentBehavior
}

// Callee 被调函数
func (c *Component) Callee(method string) reflect.Value {
	return reflect.ValueOf(c.bindMethod(method))
}

// Awake 生命周期Awake
func (c *Component) Awake() {
	method, _ := c.bindMethod("Awake").(func())
	if method != nil {
		method()
	}
}

// Start 生命周期Start
func (c *Component) Start() {
	method, _ := c.bindMethod("Start").(func())
	if method != nil {
		method()
	}
}

// Shut 生命周期Shut
func (c *Component) Shut() {
	method, _ := c.bindMethod("Shut").(func())
	if method != nil {
		method()
	}
}

// Dispose 生命周期Dispose
func (c *Component) Dispose() {
	method, _ := c.bindMethod("Dispose").(func())
	if method != nil {
		method()
	}
}

func (c *Component) bindMethod(method string) any {
	scriptPkg, ok := c.GetBuiltin().Extra.Get("script_pkg")
	if !ok {
		return nil
	}

	scriptIdent, ok := c.GetBuiltin().Extra.Get("script_ident")
	if !ok {
		return nil
	}

	thisMethod := Using(c.GetService()).Solution().BindMethod(c.GetReflected().Interface(), scriptPkg.(string), scriptIdent.(string), method)
	if thisMethod == nil {
		return nil
	}

	return thisMethod
}

// ComponentEnableUpdate 脚本化组件，支持Update
type ComponentEnableUpdate struct {
	Component
}

// Update 生命周期Update
func (c *ComponentEnableUpdate) Update() {
	method, _ := c.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

// ComponentEnableLateUpdate 脚本化组件，支持LateUpdate
type ComponentEnableLateUpdate struct {
	Component
}

// LateUpdate 生命周期LateUpdate
func (c *ComponentEnableLateUpdate) LateUpdate() {
	method, _ := c.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}

// ComponentEnableUpdateAndLateUpdate 脚本化组件，支持Update、LateUpdate
type ComponentEnableUpdateAndLateUpdate struct {
	Component
}

// Update 生命周期Update
func (c *ComponentEnableUpdateAndLateUpdate) Update() {
	method, _ := c.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

// LateUpdate 生命周期LateUpdate
func (c *ComponentEnableUpdateAndLateUpdate) LateUpdate() {
	method, _ := c.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}

// ComponentWith 创建脚本化组件原型属性，用于注册实体原型时自定义相关属性
func ComponentWith(name, script string, nonRemovable bool, extra ...map[string]any) pt.ComponentAttribute {
	return ComponentWithT[Component](name, script, nonRemovable, extra...)
}

// ComponentWithT 创建脚本化组件原型属性，用于注册实体原型时自定义相关属性
func ComponentWithT[T any](name, script string, nonRemovable bool, extra ...map[string]any) pt.ComponentAttribute {
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

	return pt.ComponentWith(types.ZeroT[T](), name, nonRemovable, _extra)
}
