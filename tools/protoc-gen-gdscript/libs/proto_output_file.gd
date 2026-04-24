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
# Buffered file-backed protobuf output stream.
# Bytes are accumulated in memory and written to FileAccess in chunk-sized batches.
class_name ProtoOutputFile
extends ProtoOutputStream

const DEFAULT_CHUNK_SIZE := 4096

var _file: FileAccess
var _chunk_size: int
var _buffer := PackedByteArray()
var _position := 0

# chunk_size controls the internal write buffer size and is clamped to 256 bytes.
func _init(file: FileAccess, chunk_size: int = DEFAULT_CHUNK_SIZE) -> void:
	if file == null:
		_init_failed = true
		_set_error(ERR_INVALID_PARAMETER, "file cannot be null")
		return
	_file = file
	_file.big_endian = false
	_chunk_size = max(chunk_size, 256)
	_buffer.resize(_chunk_size)
	_set_error(OK)

func write_byte(value: int) -> bool:
	if _init_failed:
		return false
	if !_ensure_capacity(1):
		return false
	_buffer[_position] = value & 0xFF
	_position += 1
	_set_error(OK)
	return true

func write_bytes(value: PackedByteArray) -> bool:
	if _init_failed:
		return false
	var offset := 0
	var remaining := value.size()
	if remaining <= 0:
		_set_error(OK)
		return true
	while remaining > 0:
		if !_ensure_capacity(1):
			return false
		var take := min(_chunk_size - _position, remaining)
		for i in range(take):
			_buffer[_position + i] = value[offset + i]
		_position += take
		offset += take
		remaining -= take
	_set_error(OK)
	return true

func write_fixed32(value: int) -> bool:
	if _init_failed:
		return false
	if !_ensure_capacity(4):
		return false
	_buffer.encode_u32(_position, value)
	_position += 4
	_set_error(OK)
	return true

func write_fixed64(value: int) -> bool:
	if _init_failed:
		return false
	if !_ensure_capacity(8):
		return false
	_buffer.encode_u64(_position, value)
	_position += 8
	_set_error(OK)
	return true

func write_varint(value: int) -> bool:
	if _init_failed:
		return false
	var size := ProtoUtils.sizeof_varint(value)
	if !_ensure_capacity(size):
		return false
	if value < 0:
		for i in range(9):
			_buffer[_position] = (value & 0x7F) | 0x80
			_position += 1
			value >>= 7
		_buffer[_position] = 0x01
		_position += 1
	else:
		while value >= 0x80:
			_buffer[_position] = (value & 0x7F) | 0x80
			_position += 1
			value >>= 7
		_buffer[_position] = value & 0x7F
		_position += 1
	_set_error(OK)
	return true

func write_float(value: float) -> bool:
	if _init_failed:
		return false
	if !_ensure_capacity(4):
		return false
	_buffer.encode_float(_position, value)
	_position += 4
	_set_error(OK)
	return true

func write_double(value: float) -> bool:
	if _init_failed:
		return false
	if !_ensure_capacity(8):
		return false
	_buffer.encode_double(_position, value)
	_position += 8
	_set_error(OK)
	return true

func flush() -> bool:
	if _init_failed:
		return false
	if !_flush_buffer():
		return false
	_file.flush()
	return _sync_file_error("Failed to flush file buffer.")

func _notification(what: int) -> void:
	if _init_failed:
		return
	if what == NOTIFICATION_PREDELETE:
		if !_flush_buffer():			
			push_error(
				"failed to flush on predelete, path=%s, err=%d, message=%s" % [
					_file.get_path_absolute(),
					get_error(),
					get_error_message(),
				]
			)
			return
		_file.flush()
		if !_sync_file_error("Failed to flush file buffer."):
			push_error(
				"failed to flush file buffer on predelete, path=%s, err=%d, message=%s" % [
					_file.get_path_absolute(),
					get_error(),
					get_error_message(),
				]
			)

# Flushes buffered data first when the next write would exceed the current chunk.
func _ensure_capacity(size: int) -> bool:
	if size < 0:
		_set_error(ERR_INVALID_PARAMETER, "size must be >= 0")
		return false
	if _chunk_size - _position < size:
		return _flush_buffer()
	return true

# Writes the buffered prefix to the file and resets the in-memory cursor.
func _flush_buffer() -> bool:
	if _position <= 0:
		_set_error(OK)
		return true
	if !_file.store_buffer(_buffer.slice(0, _position)):		
		return _sync_file_error("Failed to write file buffer.")
	_position = 0
	return true

func _sync_file_error(message: String) -> bool:
	var err := _file.get_error()	
	if err != OK:
		_set_error(err, message)
		return false
	_set_error(OK)
	return true
