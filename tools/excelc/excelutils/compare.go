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
	"bytes"
	"cmp"
	"math"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func ListEqual[T any](a, b []T, fun func(a, b T) bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !fun(a[i], b[i]) {
			return false
		}
	}

	return true
}

func MapEqual[K cmp.Ordered, V any](a, b map[K]V, fun func(a, b V) bool) bool {
	if len(a) != len(b) {
		return false
	}

	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if !fun(av, bv) {
			return false
		}
	}

	return true
}

func ProtoMessageEqual(a, b proto.Message) bool {
	if a == nil || b == nil {
		return false
	}
	return messageEqual(a.ProtoReflect(), b.ProtoReflect())
}

func ProtoMessageFieldsEqual(a, b protoreflect.Message, fields ...protoreflect.FieldDescriptor) bool {
	if !a.IsValid() {
		a = a.Type().Zero()
	}
	if !b.IsValid() {
		b = b.Type().Zero()
	}

	if a.Descriptor().FullName() != b.Descriptor().FullName() {
		return false
	}

	for _, fd := range fields {
		if fd == nil {
			return false
		}
		if containing := fd.ContainingMessage(); containing == nil || containing.FullName() != a.Descriptor().FullName() {
			return false
		}
		if !messageFieldEqual(fd, a.Get(fd), b.Get(fd)) {
			return false
		}
	}

	return true
}

func messageEqual(a, b protoreflect.Message) bool {
	if !a.IsValid() {
		a = a.Type().Zero()
	}
	if !b.IsValid() {
		b = b.Type().Zero()
	}

	if a.Descriptor().FullName() != b.Descriptor().FullName() {
		return false
	}

	fields := a.Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		if !messageFieldEqual(fd, a.Get(fd), b.Get(fd)) {
			return false
		}
	}

	return true
}

func messageFieldEqual(fd protoreflect.FieldDescriptor, a, b protoreflect.Value) bool {
	switch {
	case fd.IsMap():
		am := a.Map()
		bm := b.Map()

		if am.Len() != bm.Len() {
			return false
		}

		equal := true
		am.Range(func(k protoreflect.MapKey, av protoreflect.Value) bool {
			bv := bm.Get(k)
			if !bv.IsValid() || !messageFieldEqual(fd.MapValue(), av, bv) {
				equal = false
				return false
			}
			return true
		})
		return equal

	case fd.IsList():
		al := a.List()
		bl := b.List()

		if al.Len() != bl.Len() {
			return false
		}

		for i := range al.Len() {
			if !messageSingularEqual(fd, al.Get(i), bl.Get(i)) {
				return false
			}
		}
		return true

	default:
		return messageSingularEqual(fd, a, b)
	}
}

func messageSingularEqual(fd protoreflect.FieldDescriptor, a, b protoreflect.Value) bool {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return a.Bool() == b.Bool()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return a.Int() == b.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return a.Uint() == b.Uint()
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return math.Float64bits(a.Float()) == math.Float64bits(b.Float())
	case protoreflect.StringKind:
		return a.String() == b.String()
	case protoreflect.BytesKind:
		return bytes.Equal(a.Bytes(), b.Bytes())
	case protoreflect.EnumKind:
		return a.Enum() == b.Enum()
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return messageEqual(a.Message(), b.Message())
	default:
		return false
	}
}
