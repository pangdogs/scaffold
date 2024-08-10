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

func (c PropCreatorT[T]) Mem(m any) PropCreatorT[T] {
	c.ps.setMem(m)
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
