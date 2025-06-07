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

package propview

import (
	"git.golaxy.org/core/ec"
	"reflect"
)

// Deprecated: UnsafeProp 访问属性内部函数
func UnsafePropSync(ps IPropSync) _UnsafePropSync {
	return _UnsafePropSync{
		IPropSync: ps,
	}
}

type _UnsafePropSync struct {
	IPropSync
}

// Init 初始化
func (ps _UnsafePropSync) Init(view IPropView, entity ec.Entity, name string, reflectManaged reflect.Value, syncTo []string) {
	ps.init(view, entity, name, reflectManaged, syncTo)
}

// Load 加载数据
func (ps _UnsafePropSync) Load(service string) ([]byte, int64, error) {
	return ps.load(service)
}

// Save 保存数据
func (ps _UnsafePropSync) Save(service string, data []byte, revision int64) error {
	return ps.save(service, data, revision)
}

// Sync 同步变化
func (ps _UnsafePropSync) Sync(revision int64, op string, args ...any) {
	ps.sync(revision, op, args...)
}
