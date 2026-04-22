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
# A bounded view over another protobuf input stream.
# This is used for length-delimited message bodies so nested decoders cannot
# consume bytes that belong to the following field.
class_name ProtoLimitedInputStream
extends ProtoInputStream

var _stream: ProtoInputStream
var _remaining: int

# size is clamped to zero to avoid negative remaining byte counts.
func _init(stream: ProtoInputStream, size: int) -> void:
	if stream == null:
		_init_failed = true
		_set_error(ERR_INVALID_PARAMETER, "stream cannot be null")
		return
	_stream = stream
	_remaining = max(size, 0)
	_set_error(OK)

func eof() -> bool:
	if _init_failed:
		return true
	return _remaining <= 0

func read_byte() -> int:
	if _init_failed:
		return 0
	if _remaining < 1:
		_set_error(ERR_FILE_EOF, "Unexpected EOF in limited stream.")
		return 0
	var value := _stream.read_byte()
	if _stream.get_error() != OK:
		_set_error(_stream.get_error(), _stream.get_error_message())
		return 0
	_remaining -= 1
	_set_error(OK)
	return value

func read_varint() -> int:
	if _init_failed:
		return 0
	var value := 0
	var shift := 0
	while true:
		if _remaining < 1:
			_set_error(ERR_FILE_EOF, "Unexpected EOF in limited stream.")
			return 0
		var b := _stream.read_byte()
		if _stream.get_error() != OK:
			_set_error(_stream.get_error(), _stream.get_error_message())
			return 0
		_remaining -= 1
		value |= (b & 0x7F) << shift
		if (b & 0x80) == 0:
			break
		shift += 7
		if shift >= 70:
			_set_error(ERR_INVALID_DATA, "Varint is too long.")
			return 0
	_set_error(OK)
	return value

func read_bytes(size: int) -> PackedByteArray:
	if _init_failed:
		return PackedByteArray()
	if size < 0:
		_set_error(ERR_INVALID_PARAMETER, "size must be >= 0.")
		return PackedByteArray()
	if _remaining < size:
		_set_error(ERR_FILE_EOF, "Unexpected EOF in limited stream.")
		return PackedByteArray()
	var value := _stream.read_bytes(size)
	if _stream.get_error() != OK:
		_set_error(_stream.get_error(), _stream.get_error_message())
		return PackedByteArray()
	_remaining -= size
	_set_error(OK)
	return value

func read_fixed32() -> int:
	if _init_failed:
		return 0
	if _remaining < 4:
		_set_error(ERR_FILE_EOF, "Unexpected EOF in limited stream.")
		return 0
	var value := _stream.read_fixed32()
	if _stream.get_error() != OK:
		_set_error(_stream.get_error(), _stream.get_error_message())
		return 0
	_remaining -= 4
	_set_error(OK)
	return value

func read_fixed64() -> int:
	if _init_failed:
		return 0
	if _remaining < 8:
		_set_error(ERR_FILE_EOF, "Unexpected EOF in limited stream.")
		return 0
	var value := _stream.read_fixed64()
	if _stream.get_error() != OK:
		_set_error(_stream.get_error(), _stream.get_error_message())
		return 0
	_remaining -= 8
	_set_error(OK)
	return value

func read_float() -> float:
	if _init_failed:
		return 0.0
	if _remaining < 4:
		_set_error(ERR_FILE_EOF, "Unexpected EOF in limited stream.")
		return 0.0
	var value := _stream.read_float()
	if _stream.get_error() != OK:
		_set_error(_stream.get_error(), _stream.get_error_message())
		return 0.0
	_remaining -= 4
	_set_error(OK)
	return value

func read_double() -> float:
	if _init_failed:
		return 0.0
	if _remaining < 8:
		_set_error(ERR_FILE_EOF, "Unexpected EOF in limited stream.")
		return 0.0
	var value := _stream.read_double()
	if _stream.get_error() != OK:
		_set_error(_stream.get_error(), _stream.get_error_message())
		return 0.0
	_remaining -= 8
	_set_error(OK)
	return value

func skip(size: int) -> void:
	if _init_failed:
		return
	if size < 0:
		_set_error(ERR_INVALID_PARAMETER, "size must be >= 0.")
		return
	if _remaining < size:
		_set_error(ERR_FILE_EOF, "Unexpected EOF in limited stream.")
		return
	_stream.skip(size)
	if _stream.get_error() != OK:
		_set_error(_stream.get_error(), _stream.get_error_message())
		return
	_remaining -= size
	_set_error(OK)
