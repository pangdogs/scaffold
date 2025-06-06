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
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/meta"
	"reflect"
)

// IPropSync 属性同步接口
type IPropSync interface {
	IProp
	iPropSync
	// Load 加载
	Load(service string) error
	// Save 保存
	Save(service string) error
	// Sync 同步
	Sync(revision int64, op string, args ...any)
	// Managed 托管的属性
	Managed() IProp
	// Reflected 反射值
	Reflected() reflect.Value
	// Extra 额外信息
	Extra() *meta.Meta
}

type iPropSync interface {
	init(view IPropView, entity ec.Entity, name string, reflected reflect.Value, syncTo []string)
}

// PropSync 属性同步器
type PropSync struct {
	view      IPropView
	entity    ec.Entity
	name      string
	reflected reflect.Value
	syncTo    []string
	extra     generic.SliceMap[string, any]
}

func (ps *PropSync) init(view IPropView, entity ec.Entity, name string, reflected reflect.Value, syncTo []string) {
	ps.view = view
	ps.entity = entity
	ps.name = name
	ps.reflected = reflected
	ps.syncTo = syncTo
}

// Load 加载数据
func (ps *PropSync) Load(service string) ([]byte, int64, error) {
	return ps.view.Load(ps.entity.GetId(), ps.name, service)
}

// Save 保存数据
func (ps *PropSync) Save(service string, data []byte, revision int64) error {
	return ps.view.Save(ps.entity.GetId(), ps.name, service, data, revision)
}

// Sync 同步变化
func (ps *PropSync) Sync(revision int64, op string, args ...any) {
	ps.view.Sync(ps.entity.GetId(), ps.name, ps.syncTo, revision, op, args...)
}

// Reflected 反射值
func (ps *PropSync) Reflected() reflect.Value {
	return ps.reflected
}

// Extra 额外信息
func (ps *PropSync) Extra() *meta.Meta {
	return &ps.extra
}
