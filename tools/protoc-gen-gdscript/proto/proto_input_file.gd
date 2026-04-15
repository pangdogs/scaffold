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
var _source_eof := false

# chunk_size controls the internal read buffer size and is clamped to 256 bytes.
func _init(file: FileAccess, chunk_size: int = DEFAULT_CHUNK_SIZE) -> void:
	assert(file != null, "file cannot be null")	
	_file = file
	_file.big_endian = false
	_chunk_size = max(chunk_size, 256)

func eof() -> bool:
	if _available() > 0:
		return false

	_fill_buffer()
	return _available() <= 0

func read_byte() -> int:
	_ensure_available(1)
	var b: int = _buffer[_position]
	_position += 1
	_compact_buffer()
	return b

func read_bytes(size: int) -> PackedByteArray:
	assert(size >= 0, "size must be >= 0.")
	if size == 0:
		return PackedByteArray()

	var bytes := PackedByteArray()
	bytes.resize(size)

	var offset := 0
	var remaining := size

	while remaining > 0:
		if _available() <= 0:
			_fill_buffer()

		assert(_available() > 0, "Unexpected EOF while reading bytes.")

		var take: int = min(remaining, _available())
		for i in range(take):
			bytes[offset + i] = _buffer[_position + i]

		_position += take
		offset += take
		remaining -= take
		_compact_buffer()

	return bytes

func read_fixed32() -> int:
	_ensure_available(4)
	var value := _buffer.decode_u32(_position)
	_position += 4
	_compact_buffer()
	return value

func read_fixed64() -> int:
	_ensure_available(8)
	var value := _buffer.decode_u64(_position)
	_position += 8
	_compact_buffer()
	return value

func read_float() -> float:
	_ensure_available(4)
	var value := _buffer.decode_float(_position)
	_position += 4
	_compact_buffer()
	return value

func read_double() -> float:
	_ensure_available(8)
	var value := _buffer.decode_double(_position)
	_position += 8
	_compact_buffer()
	return value

# Ensures that at least size unread bytes are currently buffered.
func _ensure_available(size: int) -> void:
	assert(size >= 0, "size must be >= 0.")

	while _available() < size and not _source_eof:
		_fill_buffer()

	assert(_available() >= size, "Unexpected EOF while reading data.")

func _available() -> int:
	return _buffer.size() - _position

# Pulls another chunk from the file and appends it to the unread buffer window.
func _fill_buffer() -> void:
	if _source_eof:
		return

	_compact_buffer(true)

	var chunk := _file.get_buffer(_chunk_size)
	if chunk.is_empty():
		_source_eof = true
		return

	_buffer.append_array(chunk)

# Drops already-consumed bytes so the buffer does not grow without bound.
func _compact_buffer(force: bool = false) -> void:
	if _position <= 0:
		return

	if force or _position >= _chunk_size or _position * 2 >= _buffer.size():
		_buffer = _buffer.slice(_position)
		_position = 0
