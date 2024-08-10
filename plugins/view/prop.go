package view

import (
	"errors"
	"git.golaxy.org/framework/net/gap/variant"
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
}

type iProp interface {
	incrRevision() int64
}

// PropT 属性
type PropT[T variant.Value] struct {
	value    T
	revision int64
}

// Reset 重置
func (p *PropT[T]) Reset() {
	p.value = reflect.New(reflect.TypeFor[T]().Elem()).Interface().(T)
	p.revision = 0
}

// Value 值
func (p *PropT[T]) Value() T {
	return p.value
}

// Revision 版本号
func (p *PropT[T]) Revision() int64 {
	return p.revision
}

// Marshal 序列化
func (p *PropT[T]) Marshal() ([]byte, int64, error) {
	v, err := variant.MakeReadonlyVariant(p.value)
	if err != nil {
		return nil, 0, err
	}

	bs := make([]byte, v.Size())

	if _, err := v.Read(bs); err != nil {
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
		return errors.New("incorrect data type")
	}

	p.value = value
	p.revision = revision

	return nil
}

func (p *PropT[T]) incrRevision() int64 {
	p.revision++
	return p.revision
}