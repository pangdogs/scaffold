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

// LifecycleComponentOnCreate 组件的生命周期进入创建（OnCreate）时的回调，组件实现此接口即可使用，脚本中无法使用
type LifecycleComponentOnCreate interface {
	OnCreate()
}

// LifecycleComponentOnEnable 组件的生命周期进入启用（OnEnable）时的回调，组件实现此接口即可使用，脚本中无法使用
type LifecycleComponentOnEnable interface {
	OnEnable()
}

// LifecycleComponentOnStarted 组件的生命周期进入开始后（OnStarted）时的回调，组件实现此接口即可使用，脚本中无法使用
type LifecycleComponentOnStarted interface {
	OnStarted()
}

// LifecycleComponentOnStop 组件的生命周期进入结束前（OnStop）时的回调，组件实现此接口即可使用，脚本中无法使用
type LifecycleComponentOnStop interface {
	OnStop()
}

// LifecycleComponentOnDisable 组件的生命周期进入关闭（OnDisable）时的回调，组件实现此接口即可使用，脚本中无法使用
type LifecycleComponentOnDisable interface {
	OnDisable()
}

// LifecycleComponentOnDisposed 组件的生命周期进入释放后（OnDisposed）时的回调，组件实现此接口即可使用，脚本中无法使用
type LifecycleComponentOnDisposed interface {
	OnDisposed()
}
