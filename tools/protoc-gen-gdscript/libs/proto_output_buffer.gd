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
# In-memory protobuf output stream that accumulates bytes in a PackedByteArray.
class_name ProtoOutputBuffer
extends ProtoOutputStream

var _data := PackedByteArray()

var data: PackedByteArray:
	get:
		return _data

func _init(data: PackedByteArray = PackedByteArray()) -> void:
	_data = data
	_set_error(OK)

func write_byte(value: int) -> bool:
	_data.append(value & 0xFF)
	_set_error(OK)
	return true

func write_bytes(value: PackedByteArray) -> bool:
	if value.is_empty():
		_set_error(OK)
		return true
	_data.append_array(value)
	_set_error(OK)
	return true

func write_fixed32(value: int) -> bool:
	var offset := _data.size()
	_data.resize(offset + 4)
	_data.encode_u32(offset, value)
	_set_error(OK)
	return true

func write_fixed64(value: int) -> bool:
	var offset := _data.size()
	_data.resize(offset + 8)
	_data.encode_u64(offset, value)
	_set_error(OK)
	return true

func write_varint(value: int) -> bool:
	var size := ProtoUtils.sizeof_varint(value)
	var offset := _data.size()
	_data.resize(offset + size)
	var pos := offset
	if value < 0:
		for i in range(9):
			_data[pos] = (value & 0x7F) | 0x80
			pos += 1
			value >>= 7
		_data[pos] = 0x01
		pos += 1
	else:
		while value >= 0x80:
			_data[pos] = (value & 0x7F) | 0x80
			pos += 1
			value >>= 7
		_data[pos] = value & 0x7F
	_set_error(OK)
	return true

func write_float(value: float) -> bool:
	var offset := _data.size()
	_data.resize(offset + 4)
	_data.encode_float(offset, value)
	_set_error(OK)
	return true

func write_double(value: float) -> bool:
	var offset := _data.size()
	_data.resize(offset + 8)
	_data.encode_double(offset, value)
	_set_error(OK)
	return true

func flush() -> bool:
	_set_error(OK)
	return true
