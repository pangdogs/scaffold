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
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
	"reflect"
)

// DeclarePropT 定义属性
func DeclarePropT[T IPropSync](entity ec.Entity, name string, syncTo ...string) T {
	return declareProp(entity, name, reflect.TypeFor[T](), syncTo).(T)
}

// DeclareProp 定义属性
func DeclareProp(entity ec.Entity, name string, prop any, syncTo ...string) IPropSync {
	return declareProp(entity, name, prop, syncTo)
}

// ReferencePropT 引用属性
func ReferencePropT[T IPropSync](entity ec.Entity, name string) T {
	return referenceProp(entity, name).(T)
}

// ReferenceProp 引用属性
func ReferenceProp(entity ec.Entity, name string) IPropSync {
	return referenceProp(entity, name)
}

func declareProp(entity ec.Entity, name string, prop any, syncTo []string) IPropSync {
	if entity == nil {
		exception.Panicf("propview: %s: entity is nil", core.ErrArgs)
	}

	if prop == nil {
		exception.Panicf("propview: %s: prop is nil", core.ErrArgs)
	}

	propTab, ok := entity.(IPropTab)
	if !ok {
		exception.Panicf("propview: entity %q not implement propview.IPropTab", entity)
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
		exception.Panicf("propview: prop %q not implement propview.IPropSync", types.FullNameRT(propRT))
	}

	propInst.Managed().Reset()
	propInst.init(Using(runtime.Current(entity)), entity, name, reflect.ValueOf(propInst.Managed()), syncTo)

	propTab.AddProp(name, propInst)

	return propInst
}

func referenceProp(entity ec.Entity, name string) IPropSync {
	if entity == nil {
		exception.Panicf("propview: %s: entity is nil", core.ErrArgs)
	}

	prop := entity.(IPropTab).GetProp(name)
	if prop == nil {
		exception.Panicf("propview: prop %q not found", name)
	}

	return prop
}
