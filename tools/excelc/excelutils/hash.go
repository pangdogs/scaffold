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
	"encoding/binary"
	"errors"
	"hash"
	"hash/fnv"
	"log"
	"slices"
	"sort"

	"git.golaxy.org/core/utils/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	NewHash = fnv.New64a
)

const (
	hashTagMap   int8 = 19
	hashTagSlice int8 = 20
)

func ListToHash[T any](h hash.Hash64, l []T) error {
	if err := writeHashTag(h, hashTagSlice); err != nil {
		return err
	}
	if err := writeHashLength(h, len(l)); err != nil {
		return err
	}
	for i := range l {
		if err := AnyToHash(h, l[i]); err != nil {
			return err
		}
	}
	return nil
}

func MapToHash[K cmp.Ordered, V any](h hash.Hash64, m map[K]V) error {
	if err := writeHashTag(h, hashTagMap); err != nil {
		return err
	}
	if err := writeHashLength(h, len(m)); err != nil {
		return err
	}

	keys := make([]K, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	for _, k := range keys {
		if err := AnyToHash(h, k); err != nil {
			return err
		}
		if err := AnyToHash(h, m[k]); err != nil {
			return err
		}
	}

	return nil
}

func AnyToHash(h hash.Hash64, v any) error {
	switch iv := v.(type) {
	case bool:
		if err := writeHashTag(h, protoreflect.BoolKind); err != nil {
			return err
		}
		return writeHashByte(h, types.Bool2Int[byte](iv))
	case int32:
		if err := writeHashTag(h, protoreflect.Int32Kind); err != nil {
			return err
		}
		return binary.Write(h, binary.BigEndian, iv)
	case int64:
		if err := writeHashTag(h, protoreflect.Int64Kind); err != nil {
			return err
		}
		return binary.Write(h, binary.BigEndian, iv)
	case uint32:
		if err := writeHashTag(h, protoreflect.Uint32Kind); err != nil {
			return err
		}
		return binary.Write(h, binary.BigEndian, iv)
	case uint64:
		if err := writeHashTag(h, protoreflect.Uint64Kind); err != nil {
			return err
		}
		return binary.Write(h, binary.BigEndian, iv)
	case float32:
		if err := writeHashTag(h, protoreflect.FloatKind); err != nil {
			return err
		}
		return binary.Write(h, binary.BigEndian, iv)
	case float64:
		if err := writeHashTag(h, protoreflect.DoubleKind); err != nil {
			return err
		}
		return binary.Write(h, binary.BigEndian, iv)
	case string:
		if err := writeHashTag(h, protoreflect.StringKind); err != nil {
			return err
		}
		if err := writeHashLength(h, len(iv)); err != nil {
			return err
		}
		_, err := h.Write(types.String2Bytes(iv))
		return err
	case []byte:
		if err := writeHashTag(h, protoreflect.BytesKind); err != nil {
			return err
		}
		if err := writeHashLength(h, len(iv)); err != nil {
			return err
		}
		_, err := h.Write(iv)
		return err
	case proto.Message:
		mv := iv.ProtoReflect()
		if !mv.IsValid() {
			mv = mv.Type().Zero()
		}
		return messageToHash(h, mv)
	case protoreflect.EnumNumber:
		if err := writeHashTag(h, protoreflect.EnumKind); err != nil {
			return err
		}
		return binary.Write(h, binary.BigEndian, int32(iv))
	case protoreflect.Message:
		return messageToHash(h, iv)
	case protoreflect.List:
		return hashProtoList(h, iv)
	case protoreflect.Map:
		return hashProtoMap(h, iv)
	case protoreflect.Value:
		return AnyToHash(h, iv.Interface())
	default:
		return errors.New("value not supported for hashing")
	}
}

func messageToHash(h hash.Hash64, msg protoreflect.Message) error {
	if err := writeHashTag(h, protoreflect.MessageKind); err != nil {
		return err
	}
	if !msg.IsValid() {
		msg = msg.Type().Zero()
	}

	fields := msg.Descriptor().Fields()
	if err := writeHashLength(h, fields.Len()); err != nil {
		return err
	}

	for i := 0; i < fields.Len(); i++ {
		if err := AnyToHash(h, msg.Get(fields.Get(i))); err != nil {
			return err
		}
	}

	return nil
}

func hashProtoList(h hash.Hash64, l protoreflect.List) error {
	if err := writeHashTag(h, hashTagSlice); err != nil {
		return err
	}
	if err := writeHashLength(h, l.Len()); err != nil {
		return err
	}

	for i := 0; i < l.Len(); i++ {
		if err := AnyToHash(h, l.Get(i)); err != nil {
			return err
		}
	}

	return nil
}

func hashProtoMap(h hash.Hash64, m protoreflect.Map) error {
	if err := writeHashTag(h, hashTagMap); err != nil {
		return err
	}
	if err := writeHashLength(h, m.Len()); err != nil {
		return err
	}

	keys := make([]protoreflect.MapKey, 0, m.Len())

	m.Range(func(k protoreflect.MapKey, _ protoreflect.Value) bool {
		keys = append(keys, k)
		return true
	})

	sort.Slice(keys, func(i, j int) bool {
		return mapKeyOrder(keys[i], keys[j])
	})

	for _, k := range keys {
		if err := AnyToHash(h, k.Value()); err != nil {
			return err
		}
		if err := AnyToHash(h, m.Get(k)); err != nil {
			return err
		}
	}

	return nil
}

func writeHashTag[T ~int8](h hash.Hash64, tag T) error {
	return writeHashByte(h, byte(tag))
}

func writeHashByte(h hash.Hash64, value byte) error {
	_, err := h.Write([]byte{value})
	return err
}

func writeHashLength(h hash.Hash64, length int) error {
	if length < 0 {
		return errors.New("hash length must be non-negative")
	}
	return binary.Write(h, binary.BigEndian, uint64(length))
}

func mapKeyOrder(a, b protoreflect.MapKey) bool {
	switch a.Interface().(type) {
	case bool:
		return !a.Bool() && b.Bool()
	case int32, int64:
		return a.Int() < b.Int()
	case uint32, uint64:
		return a.Uint() < b.Uint()
	case string:
		return a.String() < b.String()
	default:
		log.Panic("invalid map key type for hashing")
	}
	panic("unreachable")
}
