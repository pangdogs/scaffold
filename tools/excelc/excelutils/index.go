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

package excelutils

import (
	"cmp"
	"errors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"math"
)

func BooleanToIndex(b bool) uint64 {
	if b {
		return 1
	} else {
		return 0
	}
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func IntegerToIndex[T Integer](v T) uint64 {
	return uint64(v)
}

func FloatToIndex(v float32) uint64 {
	return uint64(math.Float32bits(v))
}

func DoubleToIndex(v float64) uint64 {
	return math.Float64bits(v)
}

func StringToIndex(s string) (uint64, error) {
	h := NewHash()
	if err := AnyToHash(h, s); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}

func BytesToIndex(bs []byte) (uint64, error) {
	h := NewHash()
	if err := AnyToHash(h, bs); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}

func ListToIndex[T any](l []T) (uint64, error) {
	h := NewHash()
	if err := ListToHash(h, l); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}

func MapToIndex[K cmp.Ordered, V any](m map[K]V) (uint64, error) {
	h := NewHash()
	if err := MapToHash(h, m); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}

func AnyToIndex(v any) (uint64, error) {
	h := NewHash()
	if err := AnyToHash(h, v); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}

func ProtoMessageFieldToIndex(msg protoreflect.Message, field protoreflect.FieldDescriptor) (uint64, error) {
	fieldValue := msg.Get(field)
	if !fieldValue.IsValid() {
		return 0, errors.New("field is invalid")
	}

	switch v := fieldValue.Interface().(type) {
	case bool:
		return BooleanToIndex(v), nil
	case int32:
		return IntegerToIndex(v), nil
	case int64:
		return IntegerToIndex(v), nil
	case uint32:
		return IntegerToIndex(v), nil
	case uint64:
		return IntegerToIndex(v), nil
	case float32:
		return FloatToIndex(v), nil
	case float64:
		return DoubleToIndex(v), nil
	case string:
		return StringToIndex(v)
	case []byte:
		return BytesToIndex(v)
	case protoreflect.EnumNumber:
		return IntegerToIndex(v), nil
	case protoreflect.Message, protoreflect.List, protoreflect.Map:
		return AnyToIndex(fieldValue)
	default:
		return 0, errors.New("field type not supported for indexing")
	}
}

func ProtoMessageFieldNeedHashIndex(field protoreflect.FieldDescriptor) bool {
	if field.IsList() || field.IsMap() {
		return true
	}
	switch field.Kind() {
	case protoreflect.BoolKind, protoreflect.EnumKind,
		protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind, protoreflect.FloatKind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
		return false
	default:
		return true
	}
}
