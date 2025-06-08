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
	"errors"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/utils/binaryutil"
	"reflect"
)

// IProp 属性接口
type IProp interface {
	iProp
	// Reset 重置
	Reset()
	// Revision 版本号
	Revision() int64
	// Marshal 序列化
	Marshal() ([]byte, int64, error)
	// Unmarshal 反序列化
	Unmarshal(data []byte, revision int64) error
	// VariantState 可变类型状态值
	VariantState() variant.Value
	// ReflectedState 反射状态值
	ReflectedState() reflect.Value
}

// IPropStateResetCB 属性状态值重置回调
type IPropStateResetCB interface {
	OnReset()
}

type iProp interface {
	incrRevision() int64
}

// PropT 属性
type PropT[T variant.Value] struct {
	state          T
	reflectedState reflect.Value
	revision       int64
}

// Reset 重置
func (p *PropT[T]) Reset() {
	p.reflectedState = reflect.New(reflect.TypeFor[T]().Elem())
	p.state = p.reflectedState.Interface().(T)
	p.revision = 0

	if cb, ok := p.reflectedState.Interface().(IPropStateResetCB); ok {
		cb.OnReset()
	}
}

// Revision 版本号
func (p *PropT[T]) Revision() int64 {
	return p.revision
}

// Marshal 序列化
func (p *PropT[T]) Marshal() ([]byte, int64, error) {
	v, err := variant.MakeReadonlyVariant(p.state)
	if err != nil {
		return nil, 0, err
	}

	bs := make([]byte, v.Size())

	if _, err := binaryutil.CopyToBuff(bs, v); err != nil {
		return nil, 0, err
	}

	return bs, p.revision, nil
}

// Unmarshal 反序列化
func (p *PropT[T]) Unmarshal(data []byte, revision int64) error {
	v := variant.Variant{}

	if _, err := v.Write(data); err != nil {
		return err
	}

	value, ok := v.Value.(T)
	if !ok {
		return errors.New("incorrect state type")
	}

	p.state = value
	p.revision = revision

	return nil
}

// VariantState 可变类型状态值
func (p *PropT[T]) VariantState() variant.Value {
	return p.state
}

// ReflectedState 反射状态值
func (p *PropT[T]) ReflectedState() reflect.Value {
	return p.reflectedState
}

// State 状态值
func (p *PropT[T]) State() T {
	return p.state
}

func (p *PropT[T]) incrRevision() int64 {
	p.revision++
	return p.revision
}
