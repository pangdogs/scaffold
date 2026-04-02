# Abstract byte reader used by the protobuf runtime.
# Concrete implementations can read from files, memory buffers, or bounded views.
@abstract
class_name ProtoInputStream
extends RefCounted

@abstract
# Returns true when no more bytes can be read from the stream.
func eof() -> bool

@abstract
# Reads a single byte and returns it as an integer in the range [0, 255].
func read_byte() -> int

@abstract
# Reads exactly size bytes or fails if the stream ends early.
func read_bytes(size: int) -> PackedByteArray

@abstract
# Reads a little-endian 32-bit fixed-width integer.
func read_fixed32() -> int

@abstract
# Reads a little-endian 64-bit fixed-width integer.
func read_fixed64() -> int

# Reads a protobuf varint using the standard 7-bit continuation encoding.
func read_varint() -> int:
	var value := 0
	var shift := 0
	while true:
		var b := read_byte()
		value |= (b & 0x7F) << shift
		if (b & 0x80) == 0:
			break
		shift += 7
		assert(shift < 70, "Varint is too long.")
	return value
	
@abstract
# Reads a 32-bit floating-point value encoded in little-endian order.
func read_float() -> float

@abstract
# Reads a 64-bit floating-point value encoded in little-endian order.
func read_double() -> float

# Discards size bytes from the stream.
func skip(size: int) -> void:
	read_bytes(size)
