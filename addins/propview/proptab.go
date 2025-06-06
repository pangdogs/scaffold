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
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
)

// IPropTab 属性表接口
type IPropTab interface {
	// AddProp 添加属性
	AddProp(name string, ps IPropSync)
	// GetProp 获取属性
	GetProp(name string) IPropSync
	// RangeProps 遍历属性
	RangeProps(fun generic.Func2[string, IPropSync, bool])
	// EachProps 遍历属性
	EachProps(fun generic.Action2[string, IPropSync])
}

// PropTab 属性表
type PropTab generic.SliceMap[string, IPropSync]

// AddProp 添加属性
func (pt *PropTab) AddProp(name string, ps IPropSync) {
	if !pt.toSliceMap().TryAdd(name, ps) {
		exception.Panicf("propview: prop %q already exists", name)
	}
}

// GetProp 获取属性
func (pt *PropTab) GetProp(name string) IPropSync {
	return pt.toSliceMap().Value(name)
}

// RangeProps 遍历属性
func (pt *PropTab) RangeProps(fun generic.Func2[string, IPropSync, bool]) {
	for _, kv := range *pt {
		if !fun.UnsafeCall(kv.K, kv.V) {
			return
		}
	}
}

// EachProps 遍历属性
func (pt *PropTab) EachProps(fun generic.Action2[string, IPropSync]) {
	for _, kv := range *pt {
		fun.UnsafeCall(kv.K, kv.V)
	}
}

func (pt *PropTab) toSliceMap() *generic.SliceMap[string, IPropSync] {
	return (*generic.SliceMap[string, IPropSync])(pt)
}
