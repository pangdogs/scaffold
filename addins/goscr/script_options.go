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
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/scaffold/addins/goscr/dynamic"
	"time"
)

type (
	// LoadingCB 开始加载回调
	LoadingCB = generic.Action1[*dynamic.Solution]
	// LoadedCB 加载完成回调
	LoadedCB = generic.Action1[*dynamic.Solution]
)

// ScriptOptions 所有选项
type ScriptOptions struct {
	PkgRoot                              string             // 包根路径
	Projects                             []*dynamic.Project // 脚本工程列表
	AutoHotFix                           bool               // 自动热更新
	AutoHotFixLocalDetectingDelayTime    time.Duration      // 自动热更新本地脚本文件延迟更新时间
	AutoHotFixRemoteCheckingIntervalTime time.Duration      // 自动热更新远端脚本文件检测间隔时间
	LoadingCB                            LoadingCB          // 加载完成回调
	LoadedCB                             LoadedCB           // 加载完成回调
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		With.PkgRoot("").Apply(options)
		With.Projects().Apply(options)
		With.AutoHotFix(true).Apply(options)
		With.AutoHotFixLocalDetectingDelayTime(3 * time.Second).Apply(options)
		With.AutoHotFixRemoteCheckingIntervalTime(time.Minute).Apply(options)
		With.LoadingCB(nil).Apply(options)
		With.LoadedCB(nil).Apply(options)
	}
}

// PkgRoot 包根路径
func (_Option) PkgRoot(pkgRoot string) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		options.PkgRoot = pkgRoot
	}
}

// Projects 脚本工程列表
func (_Option) Projects(projects ...*dynamic.Project) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		options.Projects = projects
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

// AutoHotFixLocalDetectingDelayTime 自动热更新本地脚本文件延迟更新时间
func (_Option) AutoHotFixLocalDetectingDelayTime(d time.Duration) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		if d < 3*time.Second {
			exception.Panicf("goscr: %w: option AutoHotFixLocalDetectingDelayTime can't be set to a value less than 3 second", core.ErrArgs)
		}
		options.AutoHotFixLocalDetectingDelayTime = d
	}
}

// AutoHotFixRemoteCheckingIntervalTime 自动热更新远端脚本文件检测间隔时间
func (_Option) AutoHotFixRemoteCheckingIntervalTime(d time.Duration) option.Setting[ScriptOptions] {
	return func(options *ScriptOptions) {
		if d < 3*time.Second {
			exception.Panicf("goscr: %w: option AutoHotFixRemoteCheckingIntervalTime can't be set to a value less than 3 second", core.ErrArgs)
		}
		options.AutoHotFixRemoteCheckingIntervalTime = d
	}
}
