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

enum DictionaryKeyOrder {
	DEFAULT,
	UINT64,
}

const SIZEOF_BOOL := 1
const SIZEOF_FIXED32 := 4
const SIZEOF_FIXED64 := 8
const SIZEOF_FLOAT32 := 4
const SIZEOF_FLOAT64 := 8

const UINT32_MASK := 0xFFFFFFFF
const INT32_SIGN_BIT := 0x80000000
const UINT32_MODULUS := 0x100000000
const INT64_VALUE_MASK := 0x7FFFFFFFFFFFFFFF
const MAX_FIELD_NUMBER := 0x1FFFFFFF
const MAX_LENGTH := 0x7FFFFFFF

const HASH_TAG_DICTIONARY := 19
const HASH_TAG_ARRAY := 20

#region Boolean

# Encodes a protobuf bool as a single varint byte.
static func encode_bool(stream: ProtoOutputStream, value: bool) -> bool:
	return stream.write_byte(int(value))

# Decodes a protobuf bool varint. Zero is false; any non-zero value is true.
static func decode_bool(stream: ProtoInputStream) -> bool:
	return stream.read_varint() != 0

#endregion

#region Fixed-width Integers

# Encodes a 32-bit fixed-width integer.
static func encode_fixed32(stream: ProtoOutputStream, value: int) -> bool:
	return stream.write_fixed32(value)

# Decodes a 32-bit fixed-width integer.
static func decode_fixed32(stream: ProtoInputStream) -> int:
	return stream.read_fixed32()

# Encodes a signed 32-bit fixed-width integer.
static func encode_sfixed32(stream: ProtoOutputStream, value: int) -> bool:
	return stream.write_fixed32(value & UINT32_MASK)

# Decodes a signed 32-bit fixed-width integer.
static func decode_sfixed32(stream: ProtoInputStream) -> int:
	var value := stream.read_fixed32()
	if stream.get_error() != OK:
		return 0
	return _to_int32(value)
# Encodes a 64-bit fixed-width integer.
static func encode_fixed64(stream: ProtoOutputStream, value: int) -> bool:
	return stream.write_fixed64(value)

# Decodes a 64-bit fixed-width integer.
static func decode_fixed64(stream: ProtoInputStream) -> int:
	return stream.read_fixed64()

#endregion

#region Varint Integers

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
static func encode_varint(stream: ProtoOutputStream, value: int) -> bool:
	return stream.write_varint(value)

# Decodes an integer from a protobuf varint.
static func decode_varint(stream: ProtoInputStream) -> int:
	return stream.read_varint()

# Encodes a signed 32-bit integer after applying protobuf's truncation rules.
static func encode_int32(stream: ProtoOutputStream, value: int) -> bool:
	return stream.write_varint(_to_int32(value))

# Decodes and sign-extends the low 32 bits of a varint.
static func decode_int32(stream: ProtoInputStream) -> int:
	var value := stream.read_varint()
	if stream.get_error() != OK:
		return 0
	return _to_int32(value)

# Encodes the low 32 bits of an unsigned integer.
static func encode_uint32(stream: ProtoOutputStream, value: int) -> bool:
	return stream.write_varint(value & UINT32_MASK)

# Decodes the low 32 bits of a varint as an unsigned integer.
static func decode_uint32(stream: ProtoInputStream) -> int:
	var value := stream.read_varint()
	if stream.get_error() != OK:
		return 0
	return value & UINT32_MASK

# Enums use the same wire representation and truncation behavior as int32.
static func encode_enum(stream: ProtoOutputStream, value: int) -> bool:
	return encode_int32(stream, value)

static func decode_enum(stream: ProtoInputStream) -> int:
	return decode_int32(stream)

#endregion

#region Floating-point Numbers

# Encodes a 32-bit floating-point value.
static func encode_float(stream: ProtoOutputStream, value: float) -> bool:
	return stream.write_float(value)

# Decodes a 32-bit floating-point value.
static func decode_float(stream: ProtoInputStream) -> float:
	return stream.read_float()
# Encodes a 64-bit floating-point value.
static func encode_double(stream: ProtoOutputStream, value: float) -> bool:
	return stream.write_double(value)

# Decodes a 64-bit floating-point value.
static func decode_double(stream: ProtoInputStream) -> float:
	return stream.read_double()

#endregion

#region Strings and Bytes

# Encodes a UTF-8 string as a length-delimited protobuf field payload.
static func encode_string(stream: ProtoOutputStream, value: String) -> bool:
	var utf8_bytes := value.to_utf8_buffer()
	var size := utf8_bytes.size()
	return stream.write_varint(size) and stream.write_bytes(utf8_bytes)

# Decodes a UTF-8 string from a length-delimited protobuf field payload.
static func decode_string(stream: ProtoInputStream) -> String:
	var size := decode_length(stream)
	if stream.get_error() != OK or size <= 0:
		return ""
	var utf8_bytes := stream.read_bytes(size)
	if stream.get_error() != OK:
		return ""
	var value := utf8_bytes.get_string_from_utf8()
	return value

# Encodes a UTF-8 StringName as a length-delimited protobuf field payload.
static func encode_string_name(stream: ProtoOutputStream, value: StringName) -> bool:
	var utf8_bytes := value.to_utf8_buffer()
	var size := utf8_bytes.size()
	return stream.write_varint(size) and stream.write_bytes(utf8_bytes)

# Decodes a UTF-8 StringName from a length-delimited protobuf field payload.
static func decode_string_name(stream: ProtoInputStream) -> StringName:
	var size := decode_length(stream)
	if stream.get_error() != OK or size <= 0:
		return StringName()
	var utf8_bytes := stream.read_bytes(size)
	if stream.get_error() != OK:
		return StringName()
	var value := utf8_bytes.get_string_from_utf8()
	return StringName(value)

# Encodes a byte array as a length-delimited protobuf field payload.
static func encode_bytes(stream: ProtoOutputStream, value: PackedByteArray) -> bool:
	return stream.write_varint(value.size()) and stream.write_bytes(value)

# Decodes a byte array from a length-delimited protobuf field payload.
static func decode_bytes(stream: ProtoInputStream) -> PackedByteArray:
	var size := decode_length(stream)
	if stream.get_error() != OK or size <= 0:
		return PackedByteArray()
	return stream.read_bytes(size)

#endregion

#region Zigzag Integers

# Encodes a signed 32-bit integer using protobuf zigzag encoding.
static func encode_zigzag32(stream: ProtoOutputStream, value: int) -> bool:
	value = _to_int32(value)
	var zv := (value << 1) ^ (value >> 31)
	return stream.write_varint(zv & UINT32_MASK)

# Decodes a signed 32-bit integer using protobuf zigzag encoding.
static func decode_zigzag32(stream: ProtoInputStream) -> int:
	var zv := stream.read_varint()
	if stream.get_error() != OK:
		return 0
	zv &= UINT32_MASK
	return (zv >> 1) ^ -(zv & 1)

# Encodes a signed 64-bit integer using protobuf zigzag encoding.
static func encode_zigzag64(stream: ProtoOutputStream, value: int) -> bool:
	var zv := (value << 1) ^ (value >> 63)
	return stream.write_varint(zv)

# Decodes a signed 64-bit integer using protobuf zigzag encoding.
static func decode_zigzag64(stream: ProtoInputStream) -> int:
	var zv := stream.read_varint()
	if stream.get_error() != OK:
		return 0
	return ((zv >> 1) & INT64_VALUE_MASK) ^ -(zv & 1)

#endregion

#region Wire Format

# Decodes a protobuf length prefix and rejects values outside the int32 range.
static func decode_length(stream: ProtoInputStream) -> int:
	var size := stream.read_varint()
	if stream.get_error() != OK:
		return -1
	if size < 0 or size > MAX_LENGTH:
		stream._set_error(ERR_INVALID_DATA, "Length exceeds the protobuf 2 GiB limit.")
		return -1
	return size

# Encodes a field number and field type into a protobuf tag.
static func encode_tag(stream: ProtoOutputStream, field_number: int, field_type: int) -> bool:
	if field_number <= 0 or field_number > MAX_FIELD_NUMBER:
		stream._set_error(ERR_INVALID_PARAMETER, "Invalid protobuf field number.")
		return false
	var wire_type := ProtoFieldDescriptor.get_field_wire_type(field_type)
	if wire_type < 0:
		stream._set_error(ERR_INVALID_PARAMETER, "Invalid protobuf field type.")
		return false
	var value := (field_number << 3) | wire_type
	return stream.write_varint(value)

# Decodes a protobuf tag value from the stream.
static func decode_tag(stream: ProtoInputStream) -> int:
	var tag := stream.read_varint()
	if stream.get_error() != OK:
		return 0
	var field_number := get_tag_field_number(tag)
	if field_number <= 0 or field_number > MAX_FIELD_NUMBER:
		stream._set_error(ERR_INVALID_DATA, "Invalid protobuf field number.")
		return 0
	return tag

# Extracts the field number from an encoded protobuf tag.
static func get_tag_field_number(tag: int) -> int:
	return tag >> 3

# Extracts the wire type from an encoded protobuf tag.
static func get_tag_wire_type(tag: int) -> int:
	return tag & 0x07

# Skips one field payload according to the protobuf wire type.
static func skip_field(stream: ProtoInputStream, wire_type: int) -> bool:
	match wire_type:
		ProtoFieldDescriptor.WireType.WIRETYPE_VARINT:
			stream.read_varint()
			return stream.get_error() == OK
		ProtoFieldDescriptor.WireType.WIRETYPE_FIXED64:
			stream.skip(8)
			return stream.get_error() == OK
		ProtoFieldDescriptor.WireType.WIRETYPE_LENGTH_DELIMITED:
			var field_size := decode_length(stream)
			if stream.get_error() != OK:
				return false
			stream.skip(field_size)
			return stream.get_error() == OK
		ProtoFieldDescriptor.WireType.WIRETYPE_FIXED32:
			stream.skip(4)
			return stream.get_error() == OK
		_:
			stream._set_error(ERR_INVALID_DATA, "Invalid protobuf wire type.")
			return false

#endregion

#region Messages

# Encodes a nested message as a length-delimited protobuf payload.
static func encode_message(stream: ProtoOutputStream, msg: ProtoMessage) -> bool:
	if msg == null:
		stream._set_error(ERR_INVALID_PARAMETER, "message cannot be null")
		return false
	var size := msg.size()
	if size < 0:
		stream._set_error(ERR_INVALID_DATA, "Message size cannot be negative.")
		return false
	if !stream.write_varint(size):
		return false
	if !msg.serialize(stream):
		if stream.get_error() == OK:
			stream._set_error(ERR_INVALID_DATA, "Message serialization failed.")
		return false
	return stream.get_error() == OK

# Decodes a nested message from a buffer containing exactly the declared payload.
static func decode_message(stream: ProtoInputStream, msg: ProtoMessage) -> bool:
	if msg == null:
		stream._set_error(ERR_INVALID_PARAMETER, "message cannot be null")
		return false
	var size := decode_length(stream)
	if stream.get_error() != OK:
		return false
	var payload := stream.read_bytes(size)
	if stream.get_error() != OK:
		return false
	var payload_stream := ProtoInputBuffer.new(payload)
	if !msg.deserialize(payload_stream) or payload_stream.get_error() != OK:
		_set_nested_input_error(stream, payload_stream, "Message deserialization failed.")
		return false
	if !payload_stream.eof():
		_set_nested_input_error(stream, payload_stream, "Message payload was not fully consumed.")
		return false
	return true

#endregion

#region Dictionaries

# Returns dictionary keys sorted with the same ordering rules used by deterministic map serialization.
static func sorted_dictionary_keys(values: Dictionary, key_order: int = DictionaryKeyOrder.DEFAULT) -> Array:
	if values.is_empty():
		return []
	var keys := values.keys()
	match key_order:
		DictionaryKeyOrder.UINT64:
			keys.sort_custom(func(a, b): return _compare_u64(a, b) < 0)
		_:
			keys.sort()
	return keys

#endregion

#region Size Calculation

# Returns the encoded size of a zigzag32 payload.
static func sizeof_zigzag32(value: int) -> int:
	value = _to_int32(value)
	return sizeof_varint(((value << 1) ^ (value >> 31)) & UINT32_MASK)

# Returns the encoded size of a zigzag64 payload.
static func sizeof_zigzag64(value: int) -> int:
	return sizeof_varint((value << 1) ^ (value >> 63))

# Returns the encoded size after applying signed int32 truncation.
static func sizeof_int32(value: int) -> int:
	return sizeof_varint(_to_int32(value))

# Returns the encoded size of the low 32 unsigned bits.
static func sizeof_uint32(value: int) -> int:
	return sizeof_varint(value & UINT32_MASK)

# Returns the encoded size of a UTF-8 string payload.
static func sizeof_string(value: String) -> int:
	if value == null:
		return 0
	var utf8_bytes := value.to_utf8_buffer()
	var size := utf8_bytes.size()
	return size + sizeof_varint(size)

# Returns the encoded size of a UTF-8 StringName payload.
static func sizeof_string_name(value: StringName) -> int:
	if value == null:
		return 0
	var utf8_bytes := value.to_utf8_buffer()
	var size := utf8_bytes.size()
	return size + sizeof_varint(size)

# Returns the encoded size of a byte array payload.
static func sizeof_bytes(value: PackedByteArray) -> int:
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
	if values.is_empty() or !value_sizer.is_valid():
		return 0
	var total := 0
	for value in values:
		total += int(value_sizer.call(value))
	return total

# Returns the encoded payload size of a packed 64-bit integer array.
static func sizeof_packed_int64_array_payload(values: PackedInt64Array, value_sizer: Callable) -> int:
	if values.is_empty() or !value_sizer.is_valid():
		return 0
	var total := 0
	for value in values:
		total += int(value_sizer.call(value))
	return total

# Returns the encoded field size of a non-packed repeated field.
static func sizeof_array(values: Array, tag_size: int, value_sizer: Callable) -> int:
	if values.is_empty() or !value_sizer.is_valid():
		return 0
	var total := 0
	for value in values:
		total += tag_size + int(value_sizer.call(value))
	return total

# Returns the encoded field size of a packed repeated field.
static func sizeof_packed_array(values: Array, tag_size: int, value_sizer: Callable) -> int:
	if values.is_empty() or !value_sizer.is_valid():
		return 0
	var payload_size := sizeof_array_payload(values, value_sizer)
	return tag_size + sizeof_varint(payload_size) + payload_size

# Returns the encoded field size of a packed repeated field stored in a PackedInt64Array.
static func sizeof_packed_int64_array(values: PackedInt64Array, tag_size: int, value_sizer: Callable) -> int:
	if values.is_empty() or !value_sizer.is_valid():
		return 0
	var payload_size := sizeof_packed_int64_array_payload(values, value_sizer)
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
	if values.is_empty() or !key_sizer.is_valid():
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

#region Hashing

# Writes the fixed message field count prefix used by generated message hash_to methods.
static func hash_message_fields(hasher: ProtoHasher, field_count: int) -> void:
	hasher.write_uint64(max(field_count, 0))

# Hashes a boolean in a stable single-byte form.
static func hash_bool(hasher: ProtoHasher, value: bool) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_BOOL)
	hasher.write_byte(int(value))

# Hashes a signed 32-bit integer.
static func hash_int32(hasher: ProtoHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_INT32)
	hasher.write_int32(value)

# Hashes an unsigned 32-bit integer.
static func hash_uint32(hasher: ProtoHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_UINT32)
	hasher.write_uint32(value)

# Hashes a signed 64-bit integer.
static func hash_int64(hasher: ProtoHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_INT64)
	hasher.write_int64(value)

# Hashes an unsigned 64-bit integer.
static func hash_uint64(hasher: ProtoHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_UINT64)
	hasher.write_uint64(value)

# Hashes an enum using its signed 32-bit numeric representation.
static func hash_enum(hasher: ProtoHasher, value: int) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_ENUM)
	hasher.write_int32(value)

# Hashes a 32-bit float by using its IEEE 754 bit pattern.
static func hash_float32(hasher: ProtoHasher, value: float) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_FLOAT)
	hasher.write_uint32(_float32_bits(value))

# Hashes a 64-bit float by using its IEEE 754 bit pattern.
static func hash_float64(hasher: ProtoHasher, value: float) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_DOUBLE)
	hasher.write_uint64(_float64_bits(value))

# Hashes a string by prefixing its byte length before the UTF-8 payload.
static func hash_string(hasher: ProtoHasher, value: String) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_STRING)
	var utf8_bytes := value.to_utf8_buffer()
	var size := utf8_bytes.size()
	hasher.write_uint64(size)
	hasher.write_bytes(utf8_bytes)

# Hashes a StringName by prefixing its byte length before the UTF-8 payload.
static func hash_string_name(hasher: ProtoHasher, value: StringName) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_STRING)
	var utf8_bytes := value.to_utf8_buffer()
	var size := utf8_bytes.size()
	hasher.write_uint64(size)
	hasher.write_bytes(utf8_bytes)

# Hashes a byte array by prefixing its length before the payload.
static func hash_bytes(hasher: ProtoHasher, value: PackedByteArray) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_BYTES)
	hasher.write_uint64(value.size())
	hasher.write_bytes(value)

# Hashes a nested protobuf message, materializing a zero instance when needed.
static func hash_message(hasher: ProtoHasher, value: ProtoMessage, default_factory: Callable = Callable()) -> void:
	hasher.write_byte(ProtoFieldDescriptor.FieldType.TYPE_MESSAGE)
	var msg := value
	if msg == null and default_factory.is_valid():
		msg = default_factory.call() as ProtoMessage
	if msg != null:
		msg.hash_to(hasher)

# Hashes an array in declaration order with an element-count prefix.
static func hash_array(hasher: ProtoHasher, values: Array, value_hasher: Callable) -> void:
	hasher.write_byte(HASH_TAG_ARRAY)
	hasher.write_uint64(values.size())
	if values.is_empty() or !value_hasher.is_valid():
		return
	for value in values:
		value_hasher.call(value)

# Hashes a packed 64-bit integer array in declaration order with an element-count prefix.
static func hash_packed_int64_array(hasher: ProtoHasher, values: PackedInt64Array, value_hasher: Callable) -> void:
	hasher.write_byte(HASH_TAG_ARRAY)
	hasher.write_uint64(values.size())
	if values.is_empty() or !value_hasher.is_valid():
		return
	for value in values:
		value_hasher.call(value)

# Hashes a dictionary in sorted-key order with an entry-count prefix.
static func hash_dictionary(
	hasher: ProtoHasher,
	values: Dictionary,
	key_hasher: Callable,
	value_hasher: Callable,
	key_order: int = DictionaryKeyOrder.DEFAULT
) -> void:
	hasher.write_byte(HASH_TAG_DICTIONARY)
	hasher.write_uint64(values.size())
	if values.is_empty() or !key_hasher.is_valid() or !value_hasher.is_valid():
		return
	for key in sorted_dictionary_keys(values, key_order):
		key_hasher.call(key)
		value_hasher.call(values[key])

#endregion

#region Equality

# Compares two 32-bit floats by their IEEE 754 bit pattern.
static func equal_float32(a: float, b: float) -> bool:
	return _float32_bits(a) == _float32_bits(b)

# Compares two 64-bit floats by their IEEE 754 bit pattern.
static func equal_float64(a: float, b: float) -> bool:
	return _float64_bits(a) == _float64_bits(b)

# Compares two nested protobuf messages, materializing zero instances when requested.
static func equal_message(a: ProtoMessage, b: ProtoMessage, default_factory: Callable = Callable()) -> bool:
	if default_factory.is_valid():
		if a == null:
			a = default_factory.call() as ProtoMessage
		if b == null:
			b = default_factory.call() as ProtoMessage
	if a == null or b == null:
		return a == b
	return a.equals(b)

# Compares two arrays element by element.
static func equal_array(a: Array, b: Array, value_equal: Callable) -> bool:
	if a == b:
		return true
	if a.size() != b.size():
		return false
	if !value_equal.is_valid():
		return false
	for i in range(a.size()):
		if !bool(value_equal.call(a[i], b[i])):
			return false
	return true

# Compares two packed 64-bit integer arrays using Godot's native implementation.
static func equal_packed_int64_array(a: PackedInt64Array, b: PackedInt64Array) -> bool:
	return a == b

# Compares two dictionaries by key membership and value equality.
static func equal_dictionary(a: Dictionary, b: Dictionary, value_equal: Callable) -> bool:
	if a == b:
		return true
	if a.size() != b.size():
		return false
	if !value_equal.is_valid():
		return false
	for key in a:
		if !b.has(key):
			return false
		if !bool(value_equal.call(a[key], b[key])):
			return false
	return true

#endregion

# Marks the parent payload as invalid while preserving the nested stream error as diagnostic context.
static func _set_nested_input_error(stream: ProtoInputStream, source: ProtoInputStream, fallback_message: String) -> void:
	var message: String = "%s payload_error=%d" % [fallback_message, source.get_error()]
	var payload_message := source.get_error_message()
	if !payload_message.is_empty():
		message += ", payload_message=%s" % payload_message
	stream._set_error(ERR_INVALID_DATA, message)

# Converts the low 32 bits of an int to its signed representation.
static func _to_int32(value: int) -> int:
	value &= UINT32_MASK
	if (value & INT32_SIGN_BIT) != 0:
		return value - UINT32_MODULUS
	return value

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
