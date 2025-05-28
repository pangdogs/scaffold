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

// EntityState 脚本化实体状态
type EntityState struct {
	framework.EntityBehavior
}

// Callee 被调函数
func (e *EntityState) Callee(method string) reflect.Value {
	return reflect.ValueOf(e.bindMethod(method))
}

// Awake 生命周期唤醒（Awake）
func (e *EntityState) Awake() {
	if cb, ok := e.GetReflected().Interface().(LifecycleEntityOnCreate); ok {
		generic.CastAction0(cb.OnCreate).Call(e.GetRuntime().GetAutoRecover(), e.GetRuntime().GetReportError())
	}

	if e.GetState() != ec.EntityState_Awake {
		return
	}

	method, _ := e.bindMethod("Awake").(func())
	if method != nil {
		method()
	}
}

// Start 生命周期开始（Start）
func (e *EntityState) Start() {
	method, _ := e.bindMethod("Start").(func())
	if method != nil {
		method()
	}

	if e.GetState() != ec.EntityState_Start {
		return
	}

	if cb, ok := e.GetReflected().Interface().(LifecycleEntityOnStarted); ok {
		generic.CastAction0(cb.OnStarted).Call(e.GetRuntime().GetAutoRecover(), e.GetRuntime().GetReportError())
	}
}

// Shut 生命周期结束（Shut）
func (e *EntityState) Shut() {
	if cb, ok := e.GetReflected().Interface().(LifecycleEntityOnStop); ok {
		generic.CastAction0(cb.OnStop).Call(e.GetRuntime().GetAutoRecover(), e.GetRuntime().GetReportError())
	}

	if e.GetState() != ec.EntityState_Shut {
		return
	}

	method, _ := e.bindMethod("Shut").(func())
	if method != nil {
		method()
	}
}

// Dispose 生命周期死亡（Death）
func (e *EntityState) Dispose() {
	method, _ := e.bindMethod("Dispose").(func())
	if method != nil {
		method()
	}

	if cb, ok := e.GetReflected().Interface().(LifecycleEntityOnDisposed); ok {
		generic.CastAction0(cb.OnDisposed).Call(e.GetRuntime().GetAutoRecover(), e.GetRuntime().GetReportError())
	}
}

// EntityStateEnableUpdate 脚本化实体状态，支持帧更新（Update）
type EntityStateEnableUpdate struct {
	EntityState
}

// Update 支持帧更新（Update）
func (e *EntityStateEnableUpdate) Update() {
	method, _ := e.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

func (e *EntityState) bindMethod(method string) any {
	scriptPkg, ok := e.GetPT().Extra().Get("script_pkg")
	if !ok {
		return nil
	}

	scriptIdent, ok := e.GetPT().Extra().Get("script_ident")
	if !ok {
		return nil
	}

	thisMethod := Using(e.GetService()).Solution().BindMethod(e.GetReflected(), scriptPkg.(string), scriptIdent.(string), method)
	if thisMethod == nil {
		return nil
	}

	return thisMethod
}

// EntityStateEnableLateUpdate 脚本化实体状态，支持帧迟滞更新（Late Update）
type EntityStateEnableLateUpdate struct {
	EntityState
}

// LateUpdate 帧迟滞更新（Late Update）
func (e *EntityStateEnableLateUpdate) LateUpdate() {
	method, _ := e.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}

// EntityStateEnableUpdateAndLateUpdate 脚本化实体状态，支持帧更新（Update）、帧迟滞更新（Late Update）
type EntityStateEnableUpdateAndLateUpdate struct {
	EntityState
}

// Update 帧更新（Update）
func (e *EntityStateEnableUpdateAndLateUpdate) Update() {
	method, _ := e.bindMethod("Update").(func())
	if method != nil {
		method()
	}
}

// LateUpdate 帧迟滞更新（Late Update）
func (e *EntityStateEnableUpdateAndLateUpdate) LateUpdate() {
	method, _ := e.bindMethod("LateUpdate").(func())
	if method != nil {
		method()
	}
}
