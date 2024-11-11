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

package scr

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"github.com/traefik/yaegi/interp"
)

type (
	// LoadingCB 开始加载回调
	LoadingCB = generic.DelegateAction1[*interp.Interpreter]
	// LoadedCB 加载完成回调
	LoadedCB = generic.DelegateAction1[*interp.Interpreter]
)

// ScriptOptions 所有选项
type ScriptOptions struct {
	PathList    []string         // 脚本目录列表
	SymbolsList []interp.Exports // 导出符号列表
	AutoHotFix  bool             // 自动热更新
	LoadingCB   LoadingCB        // 加载完成回调
	LoadedCB    LoadedCB         // 加载完成回调
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		With.PathList().Apply(options)
		With.SymbolsList().Apply(options)
		With.AutoHotFix(true).Apply(options)
		With.LoadingCB(nil).Apply(options)
		With.LoadedCB(nil).Apply(options)
	}
}

// PathList 脚本目录列表
func (_Option) PathList(l ...string) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		options.PathList = l
	}
}

// SymbolsList 导出符号列表
func (_Option) SymbolsList(l ...interp.Exports) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		options.SymbolsList = l
	}
}

// AutoHotFix 自动热更新
func (_Option) AutoHotFix(b bool) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		options.AutoHotFix = b
	}
}

// LoadingCB 开始加载回调
func (_Option) LoadingCB(cb LoadingCB) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		options.LoadingCB = cb
	}
}

// LoadedCB 加载完成回调
func (_Option) LoadedCB(cb LoadedCB) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		options.LoadedCB = cb
	}
}
