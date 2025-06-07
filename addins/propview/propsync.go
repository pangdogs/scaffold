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
	"git.golaxy.org/core/utils/meta"
	"reflect"
)

// IPropSync 属性同步接口
type IPropSync interface {
	iPropSyncer
	// Load 加载
	Load(service string) error
	// Save 保存
	Save(service string) error
	// Managed 托管的属性
	Managed() IProp
	// ReflectManaged 托管的属性反射值
	ReflectManaged() reflect.Value
	// Meta meta信息
	Meta() *meta.Meta
}

type iPropSyncer interface {
	init(view IPropView, entity ec.Entity, name string, reflectManaged reflect.Value, syncTo []string)
	load(service string) ([]byte, int64, error)
	save(service string, data []byte, revision int64) error
	sync(revision int64, op string, args ...any)
}

// PropSyncer 属性同步器
type PropSyncer struct {
	view           IPropView
	entity         ec.Entity
	name           string
	reflectManaged reflect.Value
	syncTo         []string
	meta           meta.Meta
}

func (ps *PropSyncer) init(view IPropView, entity ec.Entity, name string, reflectManaged reflect.Value, syncTo []string) {
	ps.view = view
	ps.entity = entity
	ps.name = name
	ps.reflectManaged = reflectManaged
	ps.syncTo = syncTo
}

func (ps *PropSyncer) load(service string) ([]byte, int64, error) {
	return ps.view.Load(ps.entity.GetId(), ps.name, service)
}

func (ps *PropSyncer) save(service string, data []byte, revision int64) error {
	return ps.view.Save(ps.entity.GetId(), ps.name, service, data, revision)
}

func (ps *PropSyncer) sync(revision int64, op string, args ...any) {
	ps.view.Sync(ps.entity.GetId(), ps.name, ps.syncTo, revision, op, args...)
}

// ReflectManaged 托管的属性反射值
func (ps *PropSyncer) ReflectManaged() reflect.Value {
	return ps.reflectManaged
}

// Meta meta信息
func (ps *PropSyncer) Meta() *meta.Meta {
	return &ps.meta
}
