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
# Stateless protobuf encoding, sizing, and hashing helpers shared by generated message classes.
class_name ProtoUtils
extends RefCounted

const SIZEOF_BOOL := 1
const SIZEOF_FIXED32 := 4
const SIZEOF_FIXED64 := 8
const SIZEOF_FLOAT32 := 4
const SIZEOF_FLOAT64 := 8

enum DictionaryKeyOrder {
	DEFAULT,
	UINT64,
}

const HASH_TAG_DICTIONARY := 19
const HASH_TAG_ARRAY := 20

#region FNV-1a Hasher
# Incremental 64-bit FNV-1a hasher used by generated message and index code.
class Fnv64aHasher:
	const _FNV64_OFFSET_BASIS := -3750763034362895579
	const _FNV64_PRIME := 1099511628211

	var _state := _FNV64_OFFSET_BASIS

	func write_byte(value: int) -> void:
		_state = (_state ^ (value & 0xFF)) * _FNV64_PRIME

	func write_int32(value: int) -> void:
		for shift in [24, 16, 8, 0]:
			write_byte((value >> shift) & 0xFF)

	func write_uint32(value: int) -> void:
		for shift in [24, 16, 8, 0]:
			write_byte((value >> shift) & 0xFF)

	func write_int64(value: int) -> void:
		for shift in [56, 48, 40, 32, 24, 16, 8, 0]:
			write_byte((value >> shift) & 0xFF)

	func write_uint64(value: int) -> void:
		for shift in [56, 48, 40, 32, 24, 16, 8, 0]:
			write_byte((value >> shift) & 0xFF)

	func write_bytes(value: PackedByteArray) -> void:
		for byte in value:
			write_byte(byte)

	func sum64() -> int:
		return _state
#endregion

#region Bool
# Encodes a protobuf bool as a single varint byte.
static func encode_bool(stream: ProtoOutputStream, value: bool) -> void:
	stream.write_byte(int(value))

# Decodes a protobuf bool from a varint byte.
static func decode_bool(stream: ProtoInputStream) -> bool:
	return stream.read_byte() != 0
#endregion

#region Fixed32
# Encodes a 32-bit fixed-width integer.
static func encode_fixed32(stream: ProtoOutputStream, value: int) -> void:
	stream.write_fixed32(value)

# Decodes a 32-bit fixed-width integer.
static func decode_fixed32(stream: ProtoInputStream) -> int:
	return stream.read_fixed32()
#endregion

#region Fixed64
# Encodes a 64-bit fixed-width integer.
static func encode_fixed64(stream: ProtoOutputStream, value: int) -> void:
	stream.write_fixed64(value)

# Decodes a 64-bit fixed-width integer.
static func decode_fixed64(stream: ProtoInputStream) -> int:
	return stream.read_fixed64()
#endregion

#region Varint
# Returns the number of bytes required to encode value as a protobuf varint.
static func sizeof_varint(value: int) -> int:
	if value < 0:
		return 10
	var size := 1
	while value >= 0x80:
		value >>= 7
		size += 1
	return size

# Encodes an integer as a protobuf varint.
static func encode_varint(stream: ProtoOutputStream, value: int) -> void:
	stream.write_varint(value)

# Decodes an integer from a protobuf varint.
static func decode_varint(stream: ProtoInputStream) -> int:
	return stream.read_varint()
#endregion

#region Float
# Encodes a 32-bit floating-point value.
static func encode_float(stream: ProtoOutputStream, value: float) -> void:
	stream.write_float(value)

# Decodes a 32-bit floating-point value.
static func decode_float(stream: ProtoInputStream) -> float:
	return stream.read_float()
#endregion

#region Double
# Encodes a 64-bit floating-point value.
static func encode_double(stream: ProtoOutputStream, value: float) -> void:
	stream.write_double(value)

# Decodes a 64-bit floating-point value.
static func decode_double(stream: ProtoInputStream) -> float:
	return stream.read_double()
#endregion

#region String
# Encodes a UTF-8 string as a length-delimited protobuf field payload.
static func encode_string(stream: ProtoOutputStream, value) -> void:
	var utf8_value := _string_utf8_bytes(value)
	var size := utf8_value.size()
	encode_varint(stream, size)
	stream.write_bytes(utf8_value)

# Decodes a UTF-8 string from a length-delimited protobuf field payload.
static func decode_string(stream: ProtoInputStream) -> String:
	var size := decode_varint(stream)
	if size <= 0:
		return ""
	var str_bytes := stream.read_bytes(size)
	var value := str_bytes.get_string_from_utf8()
	return value

# Decodes a UTF-8 string as a StringName.
static func decode_string_name(stream: ProtoInputStream) -> StringName:
	var value := decode_string(stream)
	if value.is_empty():
		return StringName()
	return StringName(value)
#endregion

#region Bytes
# Encodes a byte array as a length-delimited protobuf field payload.
static func encode_bytes(stream: ProtoOutputStream, value: PackedByteArray) -> void:
	var size := value.size()
	encode_varint(stream, size)
	stream.write_bytes(value)

# Decodes a byte array from a length-delimited protobuf field payload.
static func decode_bytes(stream: ProtoInputStream) -> PackedByteArray:
	var size := decode_varint(stream)
	if size <= 0:
		return PackedByteArray()
	return stream.read_bytes(size)
#endregion

#region Zigzag32
# Encodes a signed 32-bit integer using protobuf zigzag encoding.
static func encode_zigzag32(stream: ProtoOutputStream, value: int) -> void:
	var zv := (value << 1) ^ (value >> 31)
	encode_varint(stream, zv)

# Decodes a signed 32-bit integer using protobuf zigzag encoding.
static func decode_zigzag32(stream: ProtoInputStream) -> int:
	var zv := decode_varint(stream)
	return (zv >> 1) ^ -(zv & 1)
#endregion

#region Zigzag64
# Encodes a signed 64-bit integer using protobuf zigzag encoding.
static func encode_zigzag64(stream: ProtoOutputStream, value: int) -> void:
	var zv := (value << 1) ^ (value >> 63)
	encode_varint(stream, zv)

# Decodes a signed 64-bit integer using protobuf zigzag encoding.
static func decode_zigzag64(stream: ProtoInputStream) -> int:
	var zv := decode_varint(stream)
	return (zv >> 1) ^ -(zv & 1)
#endregion

#region Sizeof Helpers
# Returns whether the given value should be treated as an empty string field.
static func is_empty_string(value) -> bool:
	return value == null or String(value).is_empty()

# Returns the encoded size of a zigzag32 payload.
static func sizeof_zigzag32(value: int) -> int:
	return sizeof_varint((value << 1) ^ (value >> 31))

# Returns the encoded size of a zigzag64 payload.
static func sizeof_zigzag64(value: int) -> int:
	return sizeof_varint((value << 1) ^ (value >> 63))

# Returns the encoded size of a UTF-8 string payload.
static func sizeof_string(value) -> int:
	if value == null:
		return 0
	var utf8_value := _string_utf8_bytes(value)
	var size := utf8_value.size()
	return size + sizeof_varint(size)

# Returns the encoded size of a byte array payload.
static func sizeof_bytes(value: PackedByteArray) -> int:
	if value == null:
		return 0
	var size := value.size()
	return size + sizeof_varint(size)

# Returns the encoded size of a nested message payload.
static func sizeof_message(value: ProtoMessage) -> int:
	if value == null:
		return 0
	var size := value.size()
	return size + sizeof_varint(size)

# Returns the encoded payload size of an array.
static func sizeof_array_payload(values: Array, value_sizer: Callable) -> int:
	if values == null or values.is_empty() or !value_sizer.is_valid():
		return 0
	var total := 0
	for value in values:
		total += int(value_sizer.call(value))
	return total

# Returns the encoded field size of a non-packed repeated field.
static func sizeof_array(values: Array, tag_size: int, value_sizer: Callable) -> int:
	if values == null or values.is_empty() or !value_sizer.is_valid():
		return 0
	var total := 0
	for value in values:
		total += tag_size + int(value_sizer.call(value))
	return total

# Returns the encoded field size of a packed repeated field.
static func sizeof_packed_array(values: Array, tag_size: int, value_sizer: Callable) -> int:
	if values == null or values.is_empty() or !value_sizer.is_valid():
		return 0
	var payload_size := sizeof_array_payload(values, value_sizer)
	return tag_size + sizeof_varint(payload_size) + payload_size

# Returns the encoded payload size of a single dictionary entry.
static func sizeof_dictionary_entry(
	key,
	value,
	key_tag_size: int,
	key_sizer: Callable,
	value_tag_size: int,
	value_sizer: Callable,
	value_should_serialize: Callable = Callable()
) -> int:
	if !key_sizer.is_valid():
		return 0
	var entry_size := key_tag_size + int(key_sizer.call(key))
	var should_write_value := true
	if value_should_serialize.is_valid():
		should_write_value = bool(value_should_serialize.call(value))
	if should_write_value and value_sizer.is_valid():
		entry_size += value_tag_size + int(value_sizer.call(value))
	return entry_size

# Returns the encoded field size of a dictionary field.
static func sizeof_dictionary(
	values: Dictionary,
	tag_size: int,
	key_tag_size: int,
	key_sizer: Callable,
	value_tag_size: int,
	value_sizer: Callable,
	value_should_serialize: Callable = Callable()
) -> int:
	if values == null or values.is_empty() or !key_sizer.is_valid():
		return 0
	var total := 0
	for key in values:
		var entry_size := sizeof_dictionary_entry(
			key,
			values[key],
			key_tag_size,
			key_sizer,
			value_tag_size,
			value_sizer,
			value_should_serialize
		)
		total += tag_size + sizeof_varint(entry_size) + entry_size
	return total
#endregion

#region Hash Helpers
# Allocates a new hasher for generated message and index code.
static func new_hasher() -> Fnv64aHasher:
	return Fnv64aHasher.new()

# Writes the fixed message field count prefix used by generated message hash_to methods.
static func hash_message_fields(hasher: Fnv64aHasher, field_count: int) -> void:
	hasher.write_uint64(max(field_count, 0))

# Hashes a boolean in a stable single-byte form.
static func hash_bool(hasher: Fnv64aHasher, value: bool) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_BOOL)
	hasher.write_byte(value)

# Hashes a signed 32-bit integer.
static func hash_int32(hasher: Fnv64aHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_INT32)
	hasher.write_int32(value)

# Hashes an unsigned 32-bit integer.
static func hash_uint32(hasher: Fnv64aHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_UINT32)
	hasher.write_uint32(value)

# Hashes a signed 64-bit integer.
static func hash_int64(hasher: Fnv64aHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_INT64)
	hasher.write_int64(value)

# Hashes an unsigned 64-bit integer.
static func hash_uint64(hasher: Fnv64aHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_UINT64)
	hasher.write_uint64(value)

# Hashes an enum using its signed 32-bit numeric representation.
static func hash_enum(hasher: Fnv64aHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_ENUM)
	hasher.write_int32(value)

# Hashes a 32-bit float by using its IEEE 754 bit pattern.
static func hash_float32(hasher: Fnv64aHasher, value: float) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_FLOAT)
	hasher.write_uint32(_float32_bits(value))

# Hashes a 64-bit float by using its IEEE 754 bit pattern.
static func hash_float64(hasher: Fnv64aHasher, value: float) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_DOUBLE)
	hasher.write_uint64(_float64_bits(value))

# Hashes a string by prefixing its byte length before the UTF-8 payload.
static func hash_string(hasher: Fnv64aHasher, value) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_STRING)
	var data := _string_utf8_bytes(value)
	hasher.write_uint64(data.size())
	hasher.write_bytes(data)

# Hashes a byte array by prefixing its length before the payload.
static func hash_bytes(hasher: Fnv64aHasher, value: PackedByteArray) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_BYTES)
	var data := PackedByteArray()
	if value != null:
		data = value
	hasher.write_uint64(data.size())
	hasher.write_bytes(data)

# Hashes a nested protobuf message, materializing a zero instance when needed.
static func hash_message(hasher: Fnv64aHasher, value: ProtoMessage, default_factory: Callable = Callable()) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_MESSAGE)
	var msg := value
	if msg == null and default_factory.is_valid():
		msg = default_factory.call() as ProtoMessage
	if msg != null:
		msg.hash_to(hasher)

# Hashes an array in declaration order with an element-count prefix.
static func hash_array(hasher: Fnv64aHasher, values: Array, value_hasher: Callable) -> void:
	hasher.write_byte(HASH_TAG_ARRAY)
	var size := 0
	if values != null:
		size = values.size()
	hasher.write_uint64(size)
	if values == null or values.is_empty() or !value_hasher.is_valid():
		return
	for value in values:
		value_hasher.call(value)

# Hashes a dictionary in sorted-key order with an entry-count prefix.
static func hash_dictionary(
	hasher: Fnv64aHasher,
	values: Dictionary,
	key_hasher: Callable,
	value_hasher: Callable,
	key_order: int = DictionaryKeyOrder.DEFAULT
) -> void:
	hasher.write_byte(HASH_TAG_DICTIONARY)
	var size := 0
	if values != null:
		size = values.size()
	hasher.write_uint64(size)
	if values == null or values.is_empty() or !key_hasher.is_valid() or !value_hasher.is_valid():
		return
	for key in sorted_keys(values, key_order):
		key_hasher.call(key)
		value_hasher.call(values[key])
#endregion

#region Equality Helpers
# Compares two 32-bit floats by their IEEE 754 bit pattern.
static func equal_float32(a: float, b: float) -> bool:
	return _float32_bits(a) == _float32_bits(b)

# Compares two 64-bit floats by their IEEE 754 bit pattern.
static func equal_float64(a: float, b: float) -> bool:
	return _float64_bits(a) == _float64_bits(b)

# Treats null and empty byte arrays as equivalent for equality checks.
static func equal_bytes(a, b) -> bool:
	if a == b:
		return true
	if a == null:
		return b is PackedByteArray and b.is_empty()
	if b == null:
		return a is PackedByteArray and a.is_empty()
	return false

# Compares two nested protobuf messages, materializing zero instances when requested.
static func equal_message(a: ProtoMessage, b: ProtoMessage, default_factory: Callable = Callable()) -> bool:
	var left := a
	var right := b
	if default_factory.is_valid():
		if left == null:
			left = default_factory.call() as ProtoMessage
		if right == null:
			right = default_factory.call() as ProtoMessage
	if left == null or right == null:
		return left == right
	return left.equals(right)

# Compares two arrays element by element.
static func equal_array(left: Array, right: Array, value_equal: Callable) -> bool:
	if left == right:
		return true
	if left == null:
		return right is Array and right.is_empty()
	if right == null:
		return left is Array and left.is_empty()
	if left.size() != right.size():
		return false
	if !value_equal.is_valid():
		return false
	for i in range(left.size()):
		if !bool(value_equal.call(left[i], right[i])):
			return false
	return true

# Compares two dictionaries by key membership and value equality.
static func equal_dictionary(left: Dictionary, right: Dictionary, value_equal: Callable) -> bool:
	if left == right:
		return true
	if left == null:
		return right is Dictionary and right.is_empty()
	if right == null:
		return left is Dictionary and left.is_empty()
	if left.size() != right.size():
		return false
	if !value_equal.is_valid():
		return false
	for key in left:
		if !right.has(key):
			return false
		if !bool(value_equal.call(left[key], right[key])):
			return false
	return true
#endregion

#region Tag
# Encodes a field number and field type into a protobuf tag.
static func encode_tag(stream: ProtoOutputStream, field_number: int, field_type: int) -> bool:
	if field_number <= 0:
		return false
	var wire_type := ProtoFieldDescriptor.get_wire_type(field_type)
	if wire_type < 0:
		return false
	var value := (field_number << 3) | wire_type
	encode_varint(stream, value)
	return true

# Decodes a protobuf tag value from the stream.
static func decode_tag(stream: ProtoInputStream) -> int:
	return decode_varint(stream)

# Extracts the field number from an encoded protobuf tag.
static func get_field_number(tag: int) -> int:
	return tag >> 3

# Extracts the wire type from an encoded protobuf tag.
static func get_wire_type(tag: int) -> int:
	return tag & 0x07
#endregion

#region Message
# Encodes a nested message as a length-delimited protobuf payload.
static func encode_message(stream: ProtoOutputStream, msg: ProtoMessage) -> bool:
	if msg == null:
		return false
	var size := msg.size()
	if size < 0:
		return false
	encode_varint(stream, size)
	return msg.serialize(stream)

# Decodes a nested message using a bounded substream of the declared message size.
static func decode_message(stream: ProtoInputStream, msg: ProtoMessage) -> bool:
	if msg == null:
		return false
	var size := decode_varint(stream)
	if size < 0:
		return false
	var limited_stream := ProtoLimitedInputStream.new(stream, size)
	if !msg.deserialize(limited_stream):
		return false
	return limited_stream.eof()
#endregion

#region Internal Helpers
# Returns dictionary keys sorted with the same ordering rules used by generated hashes.
static func sorted_keys(values: Dictionary, key_order: int = DictionaryKeyOrder.DEFAULT) -> Array:
	if values == null or values.is_empty():
		return []
	var keys := values.keys()
	match key_order:
		DictionaryKeyOrder.UINT64:
			keys.sort_custom(func(a, b): return _compare_u64(a, b) < 0)
		_:
			keys.sort_custom(func(a, b): return _variant_less(a, b))
	return keys

# Returns the IEEE 754 bit pattern of a 32-bit float.
static func _float32_bits(value: float) -> int:
	var buffer := PackedByteArray()
	buffer.resize(4)
	buffer.encode_float(0, value)
	return buffer.decode_u32(0)

# Returns the IEEE 754 bit pattern of a 64-bit float.
static func _float64_bits(value: float) -> int:
	var buffer := PackedByteArray()
	buffer.resize(8)
	buffer.encode_double(0, value)
	return buffer.decode_u64(0)

# Provides a stable fallback ordering for dictionary keys of mixed Variant types.
static func _variant_less(a, b) -> bool:
	match typeof(a):
		TYPE_BOOL:
			return !a and b
		TYPE_INT:
			return a < b
		TYPE_STRING:
			return String(a) < String(b)
		TYPE_STRING_NAME:
			return String(a) < String(b)
		_:
			return var_to_str(a) < var_to_str(b)

static func _string_utf8_bytes(value) -> PackedByteArray:
	if value == null:
		return PackedByteArray()
	return String(value).to_utf8_buffer()

# Compares two int values as if they were unsigned 64-bit integers.
static func _compare_u64(a: int, b: int) -> int:
	const _INT64_MIN := -9223372036854775808
	var left := a ^ _INT64_MIN
	var right := b ^ _INT64_MIN
	if left < right:
		return -1
	if left > right:
		return 1
	return 0
#endregion
