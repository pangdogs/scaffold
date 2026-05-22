# This file is part of Golaxy Distributed Service Development Framework.
#
# Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
# it under the terms of the GNU Lesser General Public License as published by
# the Free Software Foundation, either version 2.1 of the License, or
# (at your option) any later version.
#
# Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Lesser General Public License for more details.
#
# You should have received a copy of the GNU Lesser General Public License
# along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
#
# Copyright (c) 2024 pangdogs.
#
# Incremental 64-bit FNV-1a implementation used by generated proto hash code.
class_name ProtoFnv64a
extends ProtoHasher

const _FNV64_OFFSET_BASIS := -3750763034362895579
const _FNV64_PRIME := 1099511628211

var _state := _FNV64_OFFSET_BASIS

func write_byte(value: int) -> void:
	_state = (_state ^ (value & 0xFF)) * _FNV64_PRIME

func write_int32(value: int) -> void:
	for shift in [24, 16, 8, 0]:
		write_byte((value >> shift) & 0xFF)

func write_uint32(value: int) -> void:
	for shift in [24, 16, 8, 0]:
		write_byte((value >> shift) & 0xFF)

func write_int64(value: int) -> void:
	for shift in [56, 48, 40, 32, 24, 16, 8, 0]:
		write_byte((value >> shift) & 0xFF)

func write_uint64(value: int) -> void:
	for shift in [56, 48, 40, 32, 24, 16, 8, 0]:
		write_byte((value >> shift) & 0xFF)

func write_bytes(value: PackedByteArray) -> void:
	for byte in value:
		write_byte(byte)

func sum64() -> int:
	return _state
