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
# Abstract byte writer used by the protobuf runtime.
# Concrete implementations can target files, memory buffers, or custom sinks.
@abstract
class_name ProtoOutputStream
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
# Writes the least significant byte of value to the stream.
func write_byte(value: int) -> bool
	
@abstract
# Writes the raw byte array to the stream without a length prefix.
func write_bytes(value: PackedByteArray) -> bool

@abstract
# Writes a 32-bit fixed-width integer in little-endian order.
func write_fixed32(value: int) -> bool

@abstract
# Writes a 64-bit fixed-width integer in little-endian order.
func write_fixed64(value: int) -> bool

# Writes a protobuf varint using the standard 7-bit continuation encoding.
# Negative values always expand to ten bytes, matching protobuf int64 behavior.
func write_varint(value: int) -> bool:
	if _init_failed:
		return false
	_set_error(OK)
	if value < 0:
		for i in range(9):
			if !write_byte((value & 0x7F) | 0x80):
				return false
			value >>= 7
		return write_byte(0x01)
	while value >= 0x80:
		if !write_byte((value & 0x7F) | 0x80):
			return false
		value >>= 7
	return write_byte(value & 0x7F)

@abstract
# Writes a 32-bit floating-point value in little-endian order.
func write_float(value: float) -> bool

@abstract
# Writes a 64-bit floating-point value in little-endian order.
func write_double(value: float) -> bool

@abstract
# Flushes any buffered bytes to the underlying sink.
func flush() -> bool
