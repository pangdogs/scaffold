# A bounded view over another protobuf input stream.
# This is used for length-delimited message bodies so nested decoders cannot
# consume bytes that belong to the following field.
class_name ProtoLimitedInputStream
extends ProtoInputStream

var _stream: ProtoInputStream
var _remaining: int

# size is clamped to zero to avoid negative remaining byte counts.
func _init(stream: ProtoInputStream, size: int) -> void:
	assert(stream != null, "stream cannot be null")
	_stream = stream
	_remaining = max(size, 0)

func eof() -> bool:
	return _remaining <= 0

func read_byte() -> int:
	assert(_remaining >= 1, "Unexpected EOF in limited stream.")
	_remaining -= 1
	return _stream.read_byte()

func read_bytes(size: int) -> PackedByteArray:
	assert(size >= 0, "size must be >= 0.")
	assert(_remaining >= size, "Unexpected EOF in limited stream.")
	_remaining -= size
	return _stream.read_bytes(size)

func read_fixed32() -> int:
	assert(_remaining >= 4, "Unexpected EOF in limited stream.")
	_remaining -= 4
	return _stream.read_fixed32()

func read_fixed64() -> int:
	assert(_remaining >= 8, "Unexpected EOF in limited stream.")
	_remaining -= 8
	return _stream.read_fixed64()

func read_float() -> float:
	assert(_remaining >= 4, "Unexpected EOF in limited stream.")
	_remaining -= 4
	return _stream.read_float()

func read_double() -> float:
	assert(_remaining >= 8, "Unexpected EOF in limited stream.")
	_remaining -= 8
	return _stream.read_double()

func skip(size: int) -> void:
	assert(size >= 0, "size must be >= 0.")
	assert(_remaining >= size, "Unexpected EOF in limited stream.")
	_remaining -= size
	_stream.skip(size)
