extends RefCounted

const _FNV32_OFFSET_BASIS := 2166136261
const _FNV32_PRIME := 16777619
const _UINT32_MASK := 0xffffffff

var _state := _FNV32_OFFSET_BASIS

func write_byte(value: int) -> void:
	_state = ((_state ^ (value & 0xFF)) * _FNV32_PRIME) & _UINT32_MASK

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

func sum32() -> int:
	return _state & _UINT32_MASK
