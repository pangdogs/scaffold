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
# Abstract byte reader used by the protobuf runtime.
# Concrete implementations can read from files, memory buffers, or bounded views.
@abstract
class_name ProtoInputStream
extends RefCounted

var _error: int = OK
var _error_message := ""
var _init_failed := false

func get_error() -> int:
	return _error

func get_error_message() -> String:
	return _error_message

func _set_error(err: int, message: String = "") -> void:
	_error = err
	_error_message = message if err != OK else ""

@abstract
# Returns true when no more bytes can be read from the stream.
func eof() -> bool

@abstract
# Reads a single byte and returns it as an integer in the range [0, 255].
func read_byte() -> int

@abstract
# Reads exactly size bytes or fails if the stream ends early.
func read_bytes(size: int) -> PackedByteArray

@abstract
# Reads a little-endian 32-bit fixed-width integer.
func read_fixed32() -> int

@abstract
# Reads a little-endian 64-bit fixed-width integer.
func read_fixed64() -> int

# Reads a protobuf varint using the standard 7-bit continuation encoding.
func read_varint() -> int:
	if _init_failed:
		return 0
	var value := 0
	var shift := 0
	while true:
		var b := read_byte()
		if get_error() != OK:
			return 0
		value |= (b & 0x7F) << shift
		if (b & 0x80) == 0:
			break
		shift += 7
		if shift >= 70:
			_set_error(ERR_INVALID_DATA, "Varint is too long.")
			return 0
	_set_error(OK)
	return value

@abstract
# Reads a 32-bit floating-point value encoded in little-endian order.
func read_float() -> float

@abstract
# Reads a 64-bit floating-point value encoded in little-endian order.
func read_double() -> float

# Discards size bytes from the stream.
func skip(size: int) -> void:
	if _init_failed:
		return
	if size < 0:
		_set_error(ERR_INVALID_PARAMETER, "size must be >= 0.")
		return
	read_bytes(size)
