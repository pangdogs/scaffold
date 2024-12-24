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

package view

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/types"
	"reflect"
)

// PropCreatorT 属性创建器
type PropCreatorT[T IPropSync] struct {
	ps T
}

// SyncTo 设置同步目标
func (c PropCreatorT[T]) SyncTo(to ...string) PropCreatorT[T] {
	c.ps.setSyncTo(to)
	return c
}

// Atti 设置Atti
func (c PropCreatorT[T]) Atti(atti any) PropCreatorT[T] {
	c.ps.setAtti(atti)
	return c
}

// Reference 引用属性
func (c PropCreatorT[T]) Reference() T {
	return c.ps
}

// DeclarePropT 定义属性
func DeclarePropT[T IPropSync](entity ec.Entity, name string) PropCreatorT[T] {
	return PropCreatorT[T]{ps: declareProp(entity, name, reflect.TypeFor[T]()).(T)}
}

// ReferencePropT 引用属性
func ReferencePropT[T IPropSync](entity ec.Entity, name string) T {
	return referenceProp(entity, name).(T)
}

// PropCreator 属性创建器
type PropCreator = PropCreatorT[IPropSync]

// DeclareProp 定义属性
func DeclareProp(entity ec.Entity, name string, prop any) PropCreator {
	return PropCreator{ps: declareProp(entity, name, prop)}
}

// ReferenceProp 引用属性
func ReferenceProp(entity ec.Entity, name string) IPropSync {
	return referenceProp(entity, name)
}

func declareProp(entity ec.Entity, name string, prop any) IPropSync {
	if entity == nil {
		panic(fmt.Errorf("%s: entity is nil", core.ErrArgs))
	}

	if prop == nil {
		panic(fmt.Errorf("%s: prop is nil", core.ErrArgs))
	}

	propTab, ok := entity.(IPropTab)
	if !ok {
		panic(fmt.Errorf("entity %q not implement view.IPropTab", entity))
	}

	propRT, ok := prop.(reflect.Type)
	if !ok {
		propRT = reflect.TypeOf(prop)
	}

	for propRT.Kind() == reflect.Pointer {
		propRT = propRT.Elem()
	}

	propInst, ok := reflect.New(propRT).Interface().(IPropSync)
	if !ok {
		panic(fmt.Errorf("prop %q not implement view.IPropSync", types.FullNameRT(propRT)))
	}

	propInst.Reset()
	propInst.init(Using(runtime.Current(entity)), entity, name, reflect.ValueOf(propInst.Managed()))

	propTab.AddProp(name, propInst)

	return propInst
}

func referenceProp(entity ec.Entity, name string) IPropSync {
	if entity == nil {
		panic(fmt.Errorf("%s: entity is nil", core.ErrArgs))
	}

	prop := entity.(IPropTab).GetProp(name)
	if prop == nil {
		panic(fmt.Errorf("prop %s not found", name))
	}

	return prop
}
