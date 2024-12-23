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
	"reflect"
)

type PropCreatorT[T IPropSync] struct {
	ps T
}

func (c PropCreatorT[T]) SyncTo(to ...string) PropCreatorT[T] {
	c.ps.setSyncTo(to)
	return c
}

func (c PropCreatorT[T]) Atti(atti any) PropCreatorT[T] {
	c.ps.setAtti(atti)
	return c
}

func (c PropCreatorT[T]) Reference() T {
	return c.ps
}

// DeclarePropT 定义属性
func DeclarePropT[T IPropSync](entity ec.Entity, name string) PropCreatorT[T] {
	if entity == nil {
		panic(fmt.Errorf("%s: entity is nil", core.ErrArgs))
	}

	ps := reflect.New(reflect.TypeFor[T]().Elem()).Interface().(T)
	ps.Reset()
	ps.init(Using(runtime.Current(entity)), entity, name, reflect.ValueOf(ps.Managed()))

	entity.(IPropTab).AddProp(name, ps)

	return PropCreatorT[T]{ps: ps}
}

// ReferencePropT 引用属性
func ReferencePropT[T IPropSync](entity ec.Entity, name string) T {
	if entity == nil {
		panic(fmt.Errorf("%s: entity is nil", core.ErrArgs))
	}

	prop := entity.(IPropTab).GetProp(name)
	if prop == nil {
		panic(fmt.Errorf("prop %s not found", name))
	}

	return prop.(T)
}
