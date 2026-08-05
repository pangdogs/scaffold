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
# In-memory protobuf input stream backed by a PackedByteArray.
class_name ProtoInputBuffer
extends ProtoInputStream

var _data := PackedByteArray()
var _position := 0

func _init(data: PackedByteArray) -> void:
	_data = data

func eof() -> bool:
	return _position >= _data.size()

func read_byte() -> int:
	if _position >= _data.size():
		_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
		return 0
	var value := _data[_position]
	_position += 1
	if _error != OK:
		_set_error(OK)
	return value

func read_varint() -> int:
	var data_size := _data.size()
	var value := 0
	var shift := 0
	while true:
		if _position >= data_size:
			_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
			return 0
		var b := _data[_position]
		_position += 1
		if shift == 63 and b > 1:
			_set_error(ERR_INVALID_DATA, "Varint exceeds 64 bits.")
			return 0
		value |= (b & 0x7F) << shift
		if (b & 0x80) == 0:
			break
		shift += 7
	if _error != OK:
		_set_error(OK)
	return value

func read_bytes(size: int) -> PackedByteArray:
	if size < 0:
		_set_error(ERR_INVALID_PARAMETER, "size must be >= 0.")
		return PackedByteArray()
	if size == 0:
		if _error != OK:
			_set_error(OK)
		return PackedByteArray()
	if _data.size() - _position < size:
		_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
		return PackedByteArray()
	var value := _data.slice(_position, _position + size)
	_position += size
	if _error != OK:
		_set_error(OK)
	return value

func read_fixed32() -> int:
	if _data.size() - _position < 4:
		_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
		return 0
	var value := _data.decode_u32(_position)
	_position += 4
	if _error != OK:
		_set_error(OK)
	return value

func read_fixed64() -> int:
	if _data.size() - _position < 8:
		_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
		return 0
	var value := _data.decode_u64(_position)
	_position += 8
	if _error != OK:
		_set_error(OK)
	return value

func read_float() -> float:
	if _data.size() - _position < 4:
		_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
		return 0.0
	var value := _data.decode_float(_position)
	_position += 4
	if _error != OK:
		_set_error(OK)
	return value

func read_double() -> float:
	if _data.size() - _position < 8:
		_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
		return 0.0
	var value := _data.decode_double(_position)
	_position += 8
	if _error != OK:
		_set_error(OK)
	return value
