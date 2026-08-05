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
# Buffered file-backed protobuf input stream.
# Reads from FileAccess in chunks to avoid per-byte file I/O overhead.
class_name ProtoInputFile
extends ProtoInputStream

const DEFAULT_CHUNK_SIZE := 4096

var _file: FileAccess
var _chunk_size: int
var _buffer := PackedByteArray()
var _position := 0

# chunk_size controls the internal read buffer size and is clamped to 256 bytes.
func _init(file: FileAccess, chunk_size: int = DEFAULT_CHUNK_SIZE) -> void:
	if file == null:
		_init_failed = true
		_set_error(ERR_INVALID_PARAMETER, "file cannot be null")
		return
	_file = file
	_file.big_endian = false
	_chunk_size = max(chunk_size, 256)

func eof() -> bool:
	if _init_failed:
		return true
	if _position < _buffer.size():
		return false
	_fill_buffer()
	return _position >= _buffer.size()

func read_byte() -> int:
	if _init_failed:
		return 0
	if _position >= _buffer.size():
		_fill_buffer()
		if _error != OK:
			return 0
		if _position >= _buffer.size():
			_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
			return 0
	var b := _buffer[_position]
	_position += 1
	if _error != OK:
		_set_error(OK)
	return b

func read_varint() -> int:
	if _init_failed:
		return 0
	var buffer_size := _buffer.size()
	var value := 0
	var shift := 0
	while true:
		if _position >= buffer_size:
			_fill_buffer()
			if _error != OK:
				return 0
			buffer_size = _buffer.size()
			if _position >= buffer_size:
				_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
				return 0
		var b := _buffer[_position]
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
	if _init_failed:
		return PackedByteArray()
	if size < 0:
		_set_error(ERR_INVALID_PARAMETER, "size must be >= 0.")
		return PackedByteArray()
	if size == 0:
		if _error != OK:
			_set_error(OK)
		return PackedByteArray()
	var available := _buffer.size() - _position
	while available < size:
		_fill_buffer()
		if _error != OK:
			return PackedByteArray()
		var next_available := _buffer.size() - _position
		if next_available == available:
			_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
			return PackedByteArray()
		available = next_available
	var value := _buffer.slice(_position, _position + size)
	_position += size
	if _error != OK:
		_set_error(OK)
	return value

func read_fixed32() -> int:
	if _init_failed:
		return 0
	var available := _buffer.size() - _position
	while available < 4:
		_fill_buffer()
		if _error != OK:
			return 0
		var next_available := _buffer.size() - _position
		if next_available == available:
			_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
			return 0
		available = next_available
	var value := _buffer.decode_u32(_position)
	_position += 4
	if _error != OK:
		_set_error(OK)
	return value

func read_fixed64() -> int:
	if _init_failed:
		return 0
	var available := _buffer.size() - _position
	while available < 8:
		_fill_buffer()
		if _error != OK:
			return 0
		var next_available := _buffer.size() - _position
		if next_available == available:
			_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
			return 0
		available = next_available
	var value := _buffer.decode_u64(_position)
	_position += 8
	if _error != OK:
		_set_error(OK)
	return value

func read_float() -> float:
	if _init_failed:
		return 0.0
	var available := _buffer.size() - _position
	while available < 4:
		_fill_buffer()
		if _error != OK:
			return 0.0
		var next_available := _buffer.size() - _position
		if next_available == available:
			_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
			return 0.0
		available = next_available
	var value := _buffer.decode_float(_position)
	_position += 4
	if _error != OK:
		_set_error(OK)
	return value

func read_double() -> float:
	if _init_failed:
		return 0.0
	var available := _buffer.size() - _position
	while available < 8:
		_fill_buffer()
		if _error != OK:
			return 0.0
		var next_available := _buffer.size() - _position
		if next_available == available:
			_set_error(ERR_FILE_EOF, "Unexpected EOF while reading data.")
			return 0.0
		available = next_available
	var value := _buffer.decode_double(_position)
	_position += 8
	if _error != OK:
		_set_error(OK)
	return value

# Slides the unread window to the front so more file data can be appended.
func _compact_buffer(force: bool = false) -> void:
	if _position <= 0:
		return
	if force or _position >= _chunk_size or _position * 2 >= _buffer.size():
		_buffer = _buffer.slice(_position, _buffer.size())
		_position = 0

# Pulls the next chunk into the unread window.
func _fill_buffer() -> void:
	var err := _file.get_error()
	if err != OK:
		if err != ERR_FILE_EOF:
			_set_error(err, "Failed to read file buffer.")
		else:
			if _error != OK:
				_set_error(OK)
		return
	_compact_buffer(true)
	var chunk := _file.get_buffer(_chunk_size)
	err = _file.get_error()
	if err == OK or err == ERR_FILE_EOF:
		if !chunk.is_empty():
			_buffer.append_array(chunk)
	if err != OK and err != ERR_FILE_EOF:
		_set_error(err, "Failed to read file buffer.")
	else:
		if _error != OK:
			_set_error(OK)
