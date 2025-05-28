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
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework"
	"reflect"
)

// ComponentState 脚本化组件状态
type ComponentState struct {
	framework.ComponentBehavior
}

// Callee 被调函数
func (c *ComponentState) Callee(method string) reflect.Value {
	return reflect.ValueOf(c.bindMethod(method))
}

// Awake 生命周期唤醒（Awake）
func (c *ComponentState) Awake() {
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
func (c *ComponentState) OnEnable() {
	method, _ := c.bindMethod("OnEnable").(func())
	if method != nil {
		method()
	}
}

// Start 生命周期开始（Start）
func (c *ComponentState) Start() {
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
func (c *ComponentState) Shut() {
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
func (c *ComponentState) OnDisable() {
	method, _ := c.bindMethod("OnDisable").(func())
	if method != nil {
		method()
	}
}

// Dispose 生命周期死亡（Death）
func (c *ComponentState) Dispose() {
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

func (c *ComponentState) bindMethod(method string) any {
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

// ComponentStateEnableUpdate 脚本化组件状态，支持帧更新（Update）
type ComponentStateEnableUpdate struct {
	ComponentState
}

// Update 帧更新（Update）
func (c *ComponentStateEnableUpdate) Update() {
	method, _ := c.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

// ComponentStateEnableLateUpdate 脚本化组件状态，支持帧迟滞更新（Late Update）
type ComponentStateEnableLateUpdate struct {
	ComponentState
}

// LateUpdate 帧迟滞更新（Late Update）
func (c *ComponentStateEnableLateUpdate) LateUpdate() {
	method, _ := c.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}

// ComponentStateEnableUpdateAndLateUpdate 脚本化组件状态，支持帧更新（Update）、帧迟滞更新（Late Update）
type ComponentStateEnableUpdateAndLateUpdate struct {
	ComponentState
}

// Update 帧更新（Update）
func (c *ComponentStateEnableUpdateAndLateUpdate) Update() {
	method, _ := c.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

// LateUpdate 帧迟滞更新（Late Update）
func (c *ComponentStateEnableUpdateAndLateUpdate) LateUpdate() {
	method, _ := c.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}
