# In-memory protobuf output stream that accumulates bytes in a PackedByteArray.
class_name ProtoOutputBuffer
extends ProtoOutputStream

var _data := PackedByteArray()

var data: PackedByteArray:
	get:
		return _data

func write_byte(value: int) -> void:
	_data.append(value & 0xFF)

@warning_ignore("shadowed_variable")
func write_bytes(data: PackedByteArray) -> void:
	if data.is_empty():
		return
	_data.append_array(data)

func write_fixed32(value: int) -> void:
	var offset := _data.size()
	_data.resize(offset + 4)
	_data.encode_u32(offset, value)

func write_fixed64(value: int) -> void:
	var offset := _data.size()
	_data.resize(offset + 8)
	_data.encode_u64(offset, value)

func write_float(value: float) -> void:
	var offset := _data.size()
	_data.resize(offset + 4)
	_data.encode_float(offset, value)

func write_double(value: float) -> void:
	var offset := _data.size()
	_data.resize(offset + 8)
	_data.encode_double(offset, value)

func flush() -> void:
	pass
