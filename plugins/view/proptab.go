package view

import (
	"fmt"
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
		panic(fmt.Errorf("prop %q already exists", name))
	}
}

// GetProp 获取属性
func (pt *PropTab) GetProp(name string) IPropSync {
	return pt.toSliceMap().Value(name)
}

// RangeProps 遍历属性
func (pt *PropTab) RangeProps(fun generic.Func2[string, IPropSync, bool]) {
	for _, kv := range *pt {
		if !fun.Exec(kv.K, kv.V) {
			return
		}
	}
}

// EachProps 遍历属性
func (pt *PropTab) EachProps(fun generic.Action2[string, IPropSync]) {
	for _, kv := range *pt {
		fun.Exec(kv.K, kv.V)
	}
}

func (pt *PropTab) toSliceMap() *generic.SliceMap[string, IPropSync] {
	return (*generic.SliceMap[string, IPropSync])(pt)
}
