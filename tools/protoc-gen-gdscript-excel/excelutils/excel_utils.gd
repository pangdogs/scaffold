# Shared helpers used by generated Excel table access code.
# The generated code relies on this runtime for index conversion,
# dictionary key ordering, and a few small binary-search utilities.
class_name ExcelUtils
extends RefCounted

# Converts a bool into the integer form used by generated indexes.
static func boolean_to_index(value: bool) -> int:
	return int(value)

# Returns the integer index value unchanged.
static func integer_to_index(value: int) -> int:
	return value

# Converts a float into a sortable 32-bit bit-pattern index key.
static func float_to_index(value: float) -> int:
	return _float32_bits(value)

# Converts a double into a sortable 64-bit bit-pattern index key.
static func double_to_index(value: float) -> int:
	return _float64_bits(value)

# Returns dictionary keys sorted with the same ordering rules used by generated indexes.
static func sorted_keys(values: Dictionary) -> Array:
	var keys := values.keys()
	keys.sort_custom(func(a, b): return _variant_less(a, b))
	return keys

# Performs a binary search over generated index item arrays ordered by unsigned 64-bit value.
static func binary_search_index_item(items: Array, value: int) -> Variant:
	var low := 0
	var high := items.size() - 1
	while low <= high:
		@warning_ignore("integer_division")
		var mid := (low + high) / 2
		var item = items[mid]
		var cmp := _compare_u64(item.Value, value)
		if cmp < 0:
			low = mid + 1
		elif cmp > 0:
			high = mid - 1
		else:
			return item
	return null

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
		_:
			return var_to_str(a) < var_to_str(b)

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
