# In-memory protobuf input stream backed by a PackedByteArray.
class_name ProtoInputBuffer
extends ProtoInputStream

var _data := PackedByteArray()
var _position := 0

func _init(data: PackedByteArray) -> void:
	assert(data != null, "data cannot be null")
	_data = data

func eof() -> bool:
	return _position >= _data.size()

func read_byte() -> int:
	_ensure_available(1)
	var value := _data[_position]
	_position += 1
	return value

func read_bytes(size: int) -> PackedByteArray:
	assert(size >= 0, "size must be >= 0.")
	if size == 0:
		return PackedByteArray()
	_ensure_available(size)
	var value := _data.slice(_position, _position + size)
	_position += size
	return value

func read_fixed32() -> int:
	_ensure_available(4)
	var value := _data.decode_u32(_position)
	_position += 4
	return value

func read_fixed64() -> int:
	_ensure_available(8)
	var value := _data.decode_u64(_position)
	_position += 8
	return value

func read_float() -> float:
	_ensure_available(4)
	var value := _data.decode_float(_position)
	_position += 4
	return value

func read_double() -> float:
	_ensure_available(8)
	var value := _data.decode_double(_position)
	_position += 8
	return value

func _ensure_available(size: int) -> void:
	assert(size >= 0, "size must be >= 0.")
	assert(_data.size() - _position >= size, "Unexpected EOF while reading data.")
