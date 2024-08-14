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
	"git.golaxy.org/core/utils/types"
	"google.golang.org/protobuf/reflect/protoreflect"
	"hash"
	"hash/fnv"
	"slices"
	"sort"
)

var (
	NewHash = fnv.New64a
)

func ListToHash[T any](h hash.Hash64, l []T) error {
	for i := range l {
		if err := AnyToHash(h, l[i]); err != nil {
			return err
		}
	}
	return nil
}

func MapToHash[K cmp.Ordered, V any](h hash.Hash64, m map[K]V) error {
	keys := make([]K, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	for _, k := range keys {
		if err := AnyToHash(h, m[k]); err != nil {
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
	case nil:
		return nil
	case bool, int32, int64, uint32, uint64, float32, float64:
		return binary.Write(h, binary.BigEndian, iv)
	case string:
		_, err := h.Write(types.String2Bytes(iv))
		return err
	case []byte:
		_, err := h.Write(iv)
		return err
	case protoreflect.EnumNumber:
		return AnyToHash(h, int32(iv))
	case protoreflect.Message:
		for i := 0; i < iv.Descriptor().Fields().Len(); i++ {
			err := AnyToHash(h, iv.Get(iv.Descriptor().Fields().Get(i)))
			if err != nil {
				return err
			}
		}
		return nil
	case protoreflect.List:
		for i := range iv.Len() {
			if err := AnyToHash(h, iv.Get(i)); err != nil {
				return err
			}
		}
		return nil
	case protoreflect.Map:
		keys := make([]protoreflect.MapKey, 0, iv.Len())

		iv.Range(func(k protoreflect.MapKey, _ protoreflect.Value) bool {
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
			if err := AnyToHash(h, iv.Get(k)); err != nil {
				return err
			}
		}

		return nil
	case protoreflect.Value:
		return AnyToHash(h, iv.Interface())
	default:
		return errors.New("value not supported for hashing")
	}
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
		panic("invalid map key type for hashing")
	}
}
