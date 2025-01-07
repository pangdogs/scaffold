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
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework"
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

// Awake 生命周期唤醒（Awake）
func (c *Component) Awake() {
	if cb, ok := c.GetReflected().Interface().(LifecycleComponentOnCreate); ok {
		generic.CastAction0(cb.OnCreate).Call(c.GetRuntime().GetAutoRecover(), c.GetRuntime().GetReportError())
	}

	if c.GetState() != ec.ComponentState_Awake {
		return
	}

	method, _ := c.bindMethod("Awake").(func())
	if method != nil {
		method()
	}
}

// OnEnable 生命周期启用（OnEnable）
func (c *Component) OnEnable() {
	method, _ := c.bindMethod("OnEnable").(func())
	if method != nil {
		method()
	}
}

// Start 生命周期开始（Start）
func (c *Component) Start() {
	method, _ := c.bindMethod("Start").(func())
	if method != nil {
		method()
	}

	if c.GetState() != ec.ComponentState_Start {
		return
	}

	if cb, ok := c.GetReflected().Interface().(LifecycleComponentOnStarted); ok {
		generic.CastAction0(cb.OnStarted).Call(c.GetRuntime().GetAutoRecover(), c.GetRuntime().GetReportError())
	}
}

// Shut 生命周期结束（Shut）
func (c *Component) Shut() {
	if cb, ok := c.GetReflected().Interface().(LifecycleComponentOnStop); ok {
		generic.CastAction0(cb.OnStop).Call(c.GetRuntime().GetAutoRecover(), c.GetRuntime().GetReportError())
	}

	if c.GetState() != ec.ComponentState_Shut {
		return
	}

	method, _ := c.bindMethod("Shut").(func())
	if method != nil {
		method()
	}
}

// OnDisable 生命周期关闭（OnDisable）
func (c *Component) OnDisable() {
	method, _ := c.bindMethod("OnDisable").(func())
	if method != nil {
		method()
	}
}

// Dispose 生命周期死亡（Death）
func (c *Component) Dispose() {
	method, _ := c.bindMethod("Dispose").(func())
	if method != nil {
		method()
	}

	if c.GetState() != ec.ComponentState_Death {
		return
	}

	if cb, ok := c.GetReflected().Interface().(LifecycleComponentOnDisposed); ok {
		generic.CastAction0(cb.OnDisposed).Call(c.GetRuntime().GetAutoRecover(), c.GetRuntime().GetReportError())
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

	thisMethod := Using(c.GetService()).Solution().BindMethod(c.GetReflected(), scriptPkg.(string), scriptIdent.(string), method)
	if thisMethod == nil {
		return nil
	}

	return thisMethod
}

// ComponentEnableUpdate 脚本化组件，支持帧更新（Update）
type ComponentEnableUpdate struct {
	Component
}

// Update 帧更新（Update）
func (c *ComponentEnableUpdate) Update() {
	method, _ := c.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

// ComponentEnableLateUpdate 脚本化组件，支持帧迟滞更新（Late Update）
type ComponentEnableLateUpdate struct {
	Component
}

// LateUpdate 帧迟滞更新（Late Update）
func (c *ComponentEnableLateUpdate) LateUpdate() {
	method, _ := c.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}

// ComponentEnableUpdateAndLateUpdate 脚本化组件，支持帧更新（Update）、帧迟滞更新（Late Update）
type ComponentEnableUpdateAndLateUpdate struct {
	Component
}

// Update 帧更新（Update）
func (c *ComponentEnableUpdateAndLateUpdate) Update() {
	method, _ := c.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

// LateUpdate 帧迟滞更新（Late Update）
func (c *ComponentEnableUpdateAndLateUpdate) LateUpdate() {
	method, _ := c.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}

// ComponentScript 创建脚本化组件原型属性，用于注册实体原型时自定义相关属性
func ComponentScript(script string) pt.ComponentAttribute {
	return ComponentScriptT[ComponentEnableUpdateAndLateUpdate](script)
}

// ComponentScriptT 创建脚本化组件原型属性，用于注册实体原型时自定义相关属性
func ComponentScriptT[T any](script string) pt.ComponentAttribute {
	if script == "" {
		exception.Panicf("%w: script is empty", exception.ErrArgs)
	}

	idx := strings.LastIndexByte(script, '.')
	if idx < 0 {
		panic(fmt.Errorf("incorrect script %q format", script))
	}

	scriptPkg := script[:idx]
	scriptIdent := script[idx+1:]

	return pt.Component(types.ZeroT[T]()).SetName(scriptIdent).SetExtra(map[string]any{"script_pkg": scriptPkg, "script_ident": scriptIdent})
}
