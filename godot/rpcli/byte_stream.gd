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
extends RefCounted

const SIZEOF_I8 := 1
const SIZEOF_I16 := 2
const SIZEOF_I32 := 4
const SIZEOF_I64 := 8
const SIZEOF_U8 := 1
const SIZEOF_U16 := 2
const SIZEOF_U32 := 4
const SIZEOF_U64 := 8
const SIZEOF_F32 := 4
const SIZEOF_F64 := 8
const SIZEOF_BOOL := 1

var _peer: StreamPeerBuffer = StreamPeerBuffer.new()
var _error: int = OK
var _error_message: String = ""

func _init(data: PackedByteArray = PackedByteArray()) -> void:
	_peer.big_endian = true
	_peer.data_array = data.duplicate()

static func sizeof_uvarint(value: int) -> int:
	var v := max(value, 0)
	var result := 1
	while v >= 0x80:
		v >>= 7
		result += 1
	return result

static func sizeof_varint(value: int) -> int:
	var ux := value * 2 if value >= 0 else ((-value) * 2) - 1
	return sizeof_uvarint(ux)

static func sizeof_bytes(value: PackedByteArray) -> int:
	return sizeof_uvarint(value.size()) + value.size()

static func sizeof_string(value: String) -> int:
	var size := value.to_utf8_buffer().size()
	return sizeof_uvarint(size) + size

func get_error() -> int:
	return _error

func get_error_message() -> String:
	return _error_message

func clear_error() -> void:
	_error = OK
	_error_message = ""

func bytes() -> PackedByteArray:
	return _peer.data_array.duplicate()

func size() -> int:
	return _peer.data_array.size()

func position() -> int:
	return _peer.get_position()

func seek(pos: int) -> void:
	_peer.seek(pos)

func remaining() -> int:
	return _peer.get_available_bytes()

func write_raw(data: PackedByteArray) -> void:
	if get_error() != OK:
		return
	var error := _peer.put_data(data)
	if error != OK:
		_fail(error, "byte stream write failed")

func read_raw(length: int) -> PackedByteArray:
	if not _need(length):
		return PackedByteArray()
	var result: Array = _peer.get_data(length)
	if result[0] != OK:
		_fail(int(result[0]), "byte stream read failed")
		return PackedByteArray()
	return result[1]

func write_u8(value: int) -> void:
	if get_error() != OK:
		return
	_peer.put_u8(value)

func write_i8(value: int) -> void:
	if get_error() != OK:
		return
	_peer.put_8(value)

func write_u16(value: int) -> void:
	if get_error() != OK:
		return
	_peer.put_u16(value)

func write_i16(value: int) -> void:
	if get_error() != OK:
		return
	_peer.put_16(value)

func write_u32(value: int) -> void:
	if get_error() != OK:
		return
	_peer.put_u32(value)

func write_i32(value: int) -> void:
	if get_error() != OK:
		return
	_peer.put_32(value)

func write_u64(value: int) -> void:
	if get_error() != OK:
		return
	_peer.put_u64(value)

func write_i64(value: int) -> void:
	if get_error() != OK:
		return
	_peer.put_64(value)

func write_f32(value: float) -> void:
	if get_error() != OK:
		return
	_peer.put_float(value)

func write_f64(value: float) -> void:
	if get_error() != OK:
		return
	_peer.put_double(value)

func write_bool(value: bool) -> void:
	write_u8(1 if value else 0)

func write_uvarint(value: int) -> void:
	if get_error() != OK:
		return
	var v := value
	while v >= 0x80:
		write_u8((v & 0x7f) | 0x80)
		v >>= 7
	write_u8(v)

func write_varint(value: int) -> void:
	var ux := value * 2 if value >= 0 else ((-value) * 2) - 1
	write_uvarint(ux)

func write_bytes(value: PackedByteArray) -> void:
	write_uvarint(value.size())
	write_raw(value)

func write_string(value: String) -> void:
	write_bytes(value.to_utf8_buffer())

func read_u8() -> int:
	if not _need(1):
		return 0
	return _peer.get_u8()

func read_i8() -> int:
	if not _need(1):
		return 0
	return _peer.get_8()

func read_u16() -> int:
	if not _need(2):
		return 0
	return _peer.get_u16()

func read_i16() -> int:
	if not _need(2):
		return 0
	return _peer.get_16()

func read_u32() -> int:
	if not _need(4):
		return 0
	return _peer.get_u32()

func read_i32() -> int:
	if not _need(4):
		return 0
	return _peer.get_32()

func read_u64() -> int:
	if not _need(8):
		return 0
	return _peer.get_u64()

func read_i64() -> int:
	if not _need(8):
		return 0
	return _peer.get_64()

func read_f32() -> float:
	if not _need(4):
		return 0.0
	return _peer.get_float()

func read_f64() -> float:
	if not _need(8):
		return 0.0
	return _peer.get_double()

func read_bool() -> bool:
	return read_u8() != 0

func read_uvarint() -> int:
	if get_error() != OK:
		return 0
	var x := 0
	var s := 0
	for _i in range(10):
		if not _need(1):
			return 0
		var b := _peer.get_u8()
		if b < 0x80:
			x |= b << s
			return x
		x |= (b & 0x7f) << s
		s += 7
	_fail(ERR_INVALID_DATA, "invalid uvarint")
	return 0

func read_varint() -> int:
	var ux := read_uvarint()
	if get_error() != OK:
		return 0
	var x := ux >> 1
	if (ux & 1) != 0:
		x = -x - 1
	return x

func read_bytes() -> PackedByteArray:
	var length := read_uvarint()
	if get_error() != OK:
		return PackedByteArray()
	return read_raw(length)

func read_string() -> String:
	return read_bytes().get_string_from_utf8()

func _need(length: int) -> bool:
	if get_error() != OK:
		return false
	if remaining() < length:
		_fail(ERR_FILE_EOF, "unexpected eof")
		return false
	return true

func _fail(error: int, message: String) -> void:
	if _error == OK:
		_error = error
		_error_message = message
