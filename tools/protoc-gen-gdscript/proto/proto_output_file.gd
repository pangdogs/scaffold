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
	assert(file != null, "file cannot be null")
	_file = file
	_chunk_size = max(chunk_size, 256)
	_buffer.resize(_chunk_size)
	_file.big_endian = false

func write_byte(value: int) -> void:
	_ensure_capacity(1)
	_buffer[_position] = value & 0xFF
	_position += 1

func write_bytes(data: PackedByteArray) -> void:
	var offset := 0
	var remaining := data.size()

	while remaining > 0:
		if _position >= _chunk_size:
			_flush_buffer()

		var take: int = min(_chunk_size - _position, remaining)
		for i in range(take):
			_buffer[_position + i] = data[offset + i]

		_position += take
		offset += take
		remaining -= take

func write_fixed32(value: int) -> void:
	_ensure_capacity(4)
	_buffer.encode_u32(_position, value)
	_position += 4

func write_fixed64(value: int) -> void:
	_ensure_capacity(8)
	_buffer.encode_u64(_position, value)
	_position += 8

func write_float(value: float) -> void:
	_ensure_capacity(4)
	_buffer.encode_float(_position, value)
	_position += 4

func write_double(value: float) -> void:
	_ensure_capacity(8)
	_buffer.encode_double(_position, value)
	_position += 8

func flush() -> void:
	_flush_buffer()

func _notification(what: int) -> void:
	if what == NOTIFICATION_PREDELETE:
		_flush_buffer()
		_file.flush()

# Flushes buffered data first when the next write would exceed the current chunk.
func _ensure_capacity(size: int) -> void:
	if _chunk_size - _position < size:
		_flush_buffer()

# Writes the buffered prefix to the file and resets the in-memory cursor.
func _flush_buffer() -> void:
	if _position <= 0:
		return

	_file.store_buffer(_buffer.slice(0, _position))
	_position = 0
