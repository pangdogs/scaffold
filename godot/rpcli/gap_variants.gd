class_name GAPVariants
extends RefCounted

const ByteStream = preload("byte_stream.gd")

const TYPE_ID_NONE := 0
const TYPE_ID_INT := 1
const TYPE_ID_INT8 := 2
const TYPE_ID_INT16 := 3
const TYPE_ID_INT32 := 4
const TYPE_ID_INT64 := 5
const TYPE_ID_UINT := 6
const TYPE_ID_UINT8 := 7
const TYPE_ID_UINT16 := 8
const TYPE_ID_UINT32 := 9
const TYPE_ID_UINT64 := 10
const TYPE_ID_FLOAT := 11
const TYPE_ID_DOUBLE := 12
const TYPE_ID_BYTE := 13
const TYPE_ID_BOOL := 14
const TYPE_ID_BYTES := 15
const TYPE_ID_STRING := 16
const TYPE_ID_NULL := 17
const TYPE_ID_ARRAY := 18
const TYPE_ID_MAP := 19
const TYPE_ID_ERROR := 20
const TYPE_ID_CALLCHAIN := 21
const TYPE_ID_CUSTOMIZE := 32

class GAPVariant:
	extends RefCounted

	var _type_id: int = TYPE_ID_NONE
	var _value: Variant = null

	func _init(type_id: int = TYPE_ID_NONE, value: Variant = null) -> void:
		self._type_id = type_id
		self._value = value

	func type_id() -> int:
		return _type_id

	func value() ->	Variant:
		return _value

	func serialize(stream: ByteStream) -> bool:
		if !GAPVariants._write_typeid(stream, _type_id):
			return false
		return GAPVariants._write_value(stream, _type_id, _value)

	func deserialize(stream: ByteStream) -> bool:
		var variant := GAPVariants._read_variant(stream)
		if variant == null:
			return false
		_type_id = variant.type_id()
		_value = variant.value()
		return true

	func size() -> int:
		var s := GAPVariants._sizeof_value(_type_id, _value)
		if s < 0:
			return 0
		return ByteStream.SIZEOF_U32 + s

	func _to_string() -> String:
		return "GAPVariant{type_id=%d, value=%s}" % [_type_id, str(GAPVariants.to_native(self))]

class GAPArray:
	extends RefCounted

	var items: Array[GAPVariant] = []

	static func from_native(arr: Array = []) -> GAPArray:
		var ret := GAPArray.new()
		for item in arr:
			var variant := GAPVariants.to_variant(item)
			if variant == null:
				return null
			ret.items.append(variant)
		return ret

	func serialize(stream: ByteStream) -> bool:
		stream.write_uvarint(items.size())
		for item in items:
			if item == null or !item.serialize(stream):
				return false
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		var length := stream.read_uvarint()
		items.clear()
		for i in range(length):
			var item := GAPVariants.decode(stream)
			if item == null:
				return false
			items.append(item)
		return stream.get_error() == OK

	func size() -> int:
		var s := ByteStream.sizeof_uvarint(items.size())
		for item in items:
			s += item.size()
		return s

	func to_native() -> Array:
		var ret := []
		for item in items:
			ret.append(GAPVariants.to_native(item))
		return ret

	func _to_string() -> String:
		return "GAPArray%s" % [str(to_native())]

class GAPMap:
	extends RefCounted

	class KV:
		extends RefCounted

		var k: GAPVariant
		var v: GAPVariant

		func _init(k: GAPVariant, v: GAPVariant) -> void:
			self.k = k
			self.v = v

	var entries: Array[KV] = []

	static func from_native(dict: Dictionary = {}) -> GAPMap:
		var ret := GAPMap.new()
		for k in dict:
			var v = dict[k]
			var variant_k := GAPVariants.to_variant(k)
			if variant_k == null:
				return null
			var variant_v := GAPVariants.to_variant(v)
			if variant_v == null:
				return null
			ret.entries.append(KV.new(variant_k, variant_v))
		return ret

	func serialize(stream: ByteStream) -> bool:
		stream.write_uvarint(entries.size())
		for entry in entries:
			if entry == null or entry.k == null or entry.v == null:
				return false
			if !entry.k.serialize(stream) or !entry.v.serialize(stream):
				return false
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		var length := stream.read_uvarint()
		entries.clear()
		for i in range(length):
			var key := GAPVariants.decode(stream)
			if key == null:
				return false
			var value := GAPVariants.decode(stream)
			if value == null:
				return false
			entries.append(KV.new(key, value))
		return stream.get_error() == OK

	func size() -> int:
		var s := ByteStream.sizeof_uvarint(entries.size())
		for entry in entries:
			if entry == null or entry.k == null or entry.v == null:
				return 0
			s += entry.k.size() + entry.v.size()
		return s

	func to_native() -> Dictionary:
		var ret := {}
		for entry in entries:
			ret[GAPVariants.to_native(entry.k)] = GAPVariants.to_native(entry.v)
		return ret

	func _to_string() -> String:
		return "GAPMap%s" % [str(to_native())]

class GAPError:
	extends RefCounted

	var code: int = 0
	var message: String = ""

	func _init(code: int = 0, message: String = "") -> void:
		self.code = code
		self.message = message

	func serialize(stream: ByteStream) -> bool:
		stream.write_i32(code)
		stream.write_string(message)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		code = stream.read_i32()
		message = stream.read_string()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.SIZEOF_I32 + ByteStream.sizeof_string(message)

	func _to_string() -> String:
		return "GAPError{code=%d, message=%s}" % [code, message]

class GAPCall:
	extends RefCounted

	var svc: String = ""
	var addr: String = ""
	var timestamp: int = 0
	var transit: bool = false

	func _init(svc: String = "", addr: String = "", timestamp: int = 0, transit: bool = false) -> void:
		self.svc = svc
		self.addr = addr
		self.timestamp = timestamp
		self.transit = transit

	func serialize(stream: ByteStream) -> bool:
		stream.write_string(svc)
		stream.write_string(addr)
		stream.write_i64(timestamp)
		stream.write_bool(transit)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		svc = stream.read_string()
		addr = stream.read_string()
		timestamp = stream.read_i64()
		transit = stream.read_bool()
		return stream.get_error() == OK

	func size() -> int:
		return (
			ByteStream.sizeof_string(svc) +
			ByteStream.sizeof_string(addr) +
			ByteStream.SIZEOF_I64 +
			ByteStream.SIZEOF_BOOL
		)

	func _to_string() -> String:
		return "GAPCall{svc=%s, addr=%s, timestamp=%d, transit=%s}" % [svc, addr, timestamp, transit]

class GAPCallChain:
	extends RefCounted

	var items: Array[GAPCall] = []

	func _init(items: Array[GAPCall] = []) -> void:
		self.items = items

	func serialize(stream: ByteStream) -> bool:
		stream.write_uvarint(items.size())
		for item in items:
			if !item.serialize(stream):
				return false
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		var length = stream.read_uvarint()
		items.clear()
		for i in range(length):
			var item := GAPCall.new()
			if !item.deserialize(stream):
				return false
			items.append(item)
		return stream.get_error() == OK

	func size() -> int:
		var s := ByteStream.sizeof_uvarint(items.size())
		for item in items:
			s += item.size()
		return s

	func _to_string() -> String:
		return "GAPCallChain%s" % [str(items)]

class CustomTypeRegistry:
	var _types: Dictionary[int, Variant] = {}

	func register(type_id: int, type_cls: Variant) -> void:
		if type_id < TYPE_ID_CUSTOMIZE:
			push_error("GAP variant custom type_id must be >= %d" % TYPE_ID_CUSTOMIZE)
			return
		if typeof(type_cls) != TYPE_OBJECT or type_cls == null or !type_cls.has_method("new"):
			push_error("GAP variant custom type_cls is invalid, type_id=%d" % type_id)
			return
		_types[type_id] = type_cls

	func is_registered(type_id: int) -> bool:
		return _types.has(type_id)

	func new_value(type_id: int) -> Object:
		var type_cls := _types.get(type_id)
		if type_cls == null:
			return null
		return type_cls.new()

static var custom_type_registry := CustomTypeRegistry.new()

static func to_variant(value: Variant) -> GAPVariant:
	match typeof(value):
		TYPE_NIL:
			return GAPVariant.new(TYPE_ID_NULL)
		TYPE_BOOL:
			return GAPVariant.new(TYPE_ID_BOOL, value)
		TYPE_INT:
			return GAPVariant.new(TYPE_ID_INT64, value)
		TYPE_FLOAT:
			return GAPVariant.new(TYPE_ID_DOUBLE, value)
		TYPE_STRING, TYPE_STRING_NAME:
			return GAPVariant.new(TYPE_ID_STRING, value)
		TYPE_OBJECT:
			if value is GAPVariant:
				return value
			elif value is GAPArray:
				return GAPVariant.new(TYPE_ID_ARRAY, value)
			elif value is GAPMap:
				return GAPVariant.new(TYPE_ID_MAP, value)
			elif value is GAPError:
				return GAPVariant.new(TYPE_ID_ERROR, value)
			elif value is GAPCallChain:
				return GAPVariant.new(TYPE_ID_CALLCHAIN, value)
			elif value is ProtoGAPVariant:
				var type_id: int = value.type_id()
				if !custom_type_registry.is_registered(type_id):
					push_error("GAP variant custom type_id is not registered, type_id=%d" % type_id)
					return null
				return GAPVariant.new(type_id, value)
			else:
				return null
		TYPE_DICTIONARY:
			var map := GAPMap.from_native(value)
			if map == null:
				return null
			return GAPVariant.new(TYPE_ID_MAP, map)
		TYPE_ARRAY:
			var arr := GAPArray.from_native(value)
			if arr == null:
				return null
			return GAPVariant.new(TYPE_ID_ARRAY, arr)
		TYPE_PACKED_BYTE_ARRAY:
			return GAPVariant.new(TYPE_ID_BYTES, value)
	return null

static func to_native(variant: GAPVariant) -> Variant:
	if variant == null:
		return null
	match variant.type_id():
		TYPE_ID_ARRAY:
			var arr := variant.value() as GAPArray
			if arr == null:
				return null
			return arr.to_native()
		TYPE_ID_MAP:
			var map := variant.value() as GAPMap
			if map == null:
				return null
			return map.to_native()
		TYPE_ID_NULL, TYPE_ID_NONE:
			return null
		_:
			return variant.value()

static func encode(stream: ByteStream, value: Variant) -> bool:
	var variant := to_variant(value)
	if variant == null:
		return false
	return variant.serialize(stream)

static func decode(stream: ByteStream) -> GAPVariant:
	return _read_variant(stream)

static func _write_typeid(stream: ByteStream, type_id: int) -> bool:
	if type_id == TYPE_ID_NONE:
		return false
	stream.write_u32(type_id)
	return stream.get_error() == OK

static func _write_value(stream: ByteStream, type_id: int, value: Variant) -> bool:
	match type_id:
		TYPE_ID_NONE:
			return false
		TYPE_ID_INT:
			stream.write_varint(int(value))
		TYPE_ID_INT8:
			stream.write_i8(int(value))
		TYPE_ID_INT16:
			stream.write_i16(int(value))
		TYPE_ID_INT32:
			stream.write_i32(int(value))
		TYPE_ID_INT64:
			stream.write_varint(int(value))
		TYPE_ID_UINT:
			stream.write_uvarint(int(value))
		TYPE_ID_UINT8, TYPE_ID_BYTE:
			stream.write_u8(int(value))
		TYPE_ID_UINT16:
			stream.write_u16(int(value))
		TYPE_ID_UINT32:
			stream.write_u32(int(value))
		TYPE_ID_UINT64:
			stream.write_uvarint(int(value))
		TYPE_ID_FLOAT:
			stream.write_f32(float(value))
		TYPE_ID_DOUBLE:
			stream.write_f64(float(value))
		TYPE_ID_BOOL:
			stream.write_bool(bool(value))
		TYPE_ID_BYTES:
			stream.write_bytes(PackedByteArray(value))
		TYPE_ID_STRING:
			stream.write_string(String(value))
		TYPE_ID_NULL:
			return true
		TYPE_ID_ARRAY:
			var arr := value as GAPArray
			return arr != null and arr.serialize(stream)
		TYPE_ID_MAP:
			var map := value as GAPMap
			return map != null and map.serialize(stream)
		TYPE_ID_ERROR:
			var err := value as GAPError
			return err != null and err.serialize(stream)
		TYPE_ID_CALLCHAIN:
			var cc := value as GAPCallChain
			return cc != null and cc.serialize(stream)
		_:
			var obj := value as ProtoGAPVariant
			if obj == null:
				return false
			var proto_stream := ProtoOutputBuffer.new()
			if !obj.serialize(proto_stream):
				return false
			stream.write_bytes(proto_stream.data)
	return stream.get_error() == OK

static func _read_variant(stream: ByteStream) -> GAPVariant:
	var type_id := stream.read_u32()
	if stream.get_error() != OK:
		return null
	match type_id:
		TYPE_ID_NONE:
			return null
		TYPE_ID_INT:
			var value := stream.read_varint()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_INT8:
			var value := stream.read_i8()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_INT16:
			var value := stream.read_i16()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_INT32:
			var value := stream.read_i32()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_INT64:
			var value := stream.read_varint()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_UINT:
			var value := stream.read_uvarint()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_UINT8, TYPE_ID_BYTE:
			var value := stream.read_u8()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_UINT16:
			var value := stream.read_u16()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_UINT32:
			var value := stream.read_u32()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_UINT64:
			var value := stream.read_uvarint()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_FLOAT:
			var value := stream.read_f32()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_DOUBLE:
			var value := stream.read_f64()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_BOOL:
			var value := stream.read_bool()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_BYTES:
			var value := stream.read_bytes()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_STRING:
			var value := stream.read_string()
			if stream.get_error() != OK:
				return null
			return GAPVariant.new(type_id, value)
		TYPE_ID_NULL:
			return GAPVariant.new(type_id)
		TYPE_ID_ARRAY:
			var arr := GAPArray.new()
			if !arr.deserialize(stream):
				return null
			return GAPVariant.new(type_id, arr)
		TYPE_ID_MAP:
			var map := GAPMap.new()
			if !map.deserialize(stream):
				return null
			return GAPVariant.new(type_id, map)
		TYPE_ID_ERROR:
			var err := GAPError.new()
			if !err.deserialize(stream):
				return null
			return GAPVariant.new(type_id, err)
		TYPE_ID_CALLCHAIN:
			var cc := GAPCallChain.new()
			if !cc.deserialize(stream):
				return null
			return GAPVariant.new(type_id, cc)
		_:
			var obj := custom_type_registry.new_value(type_id) as ProtoGAPVariant
			if obj == null:
				return null
			var data := stream.read_bytes()
			if stream.get_error() != OK:
				return null
			var proto_stream := ProtoInputBuffer.new(data)
			if !obj.deserialize(proto_stream):
				return null
			return GAPVariant.new(type_id, obj)

static func _sizeof_value(type_id: int, value: Variant) -> int:
	match type_id:
		TYPE_ID_NONE:
			return -1
		TYPE_ID_INT:
			return ByteStream.sizeof_varint(int(value))
		TYPE_ID_INT8:
			return ByteStream.SIZEOF_I8
		TYPE_ID_INT16:
			return ByteStream.SIZEOF_I16
		TYPE_ID_INT32:
			return ByteStream.SIZEOF_I32
		TYPE_ID_INT64:
			return ByteStream.sizeof_varint(int(value))
		TYPE_ID_UINT:
			return ByteStream.sizeof_uvarint(int(value))
		TYPE_ID_UINT8, TYPE_ID_BYTE:
			return ByteStream.SIZEOF_U8
		TYPE_ID_UINT16:
			return ByteStream.SIZEOF_U16
		TYPE_ID_UINT32:
			return ByteStream.SIZEOF_U32
		TYPE_ID_UINT64:
			return ByteStream.sizeof_uvarint(int(value))
		TYPE_ID_FLOAT:
			return ByteStream.SIZEOF_F32
		TYPE_ID_DOUBLE:
			return ByteStream.SIZEOF_F64
		TYPE_ID_BOOL:
			return ByteStream.SIZEOF_BOOL
		TYPE_ID_BYTES:
			return ByteStream.sizeof_bytes(PackedByteArray(value))
		TYPE_ID_STRING:
			return ByteStream.sizeof_string(String(value))
		TYPE_ID_NULL:
			return 0
		TYPE_ID_ARRAY:
			var arr := value as GAPArray
			if arr == null:
				return -1
			return arr.size()
		TYPE_ID_MAP:
			var map := value as GAPMap
			if map == null:
				return -1
			return map.size()
		TYPE_ID_ERROR:
			var err := value as GAPError
			if err == null:
				return -1
			return err.size()
		TYPE_ID_CALLCHAIN:
			var cc := value as GAPCallChain
			if cc == null:
				return -1
			return cc.size()
		_:
			var obj := value as ProtoGAPVariant
			if obj == null:
				return -1
			var s := obj.size()
			return ByteStream.sizeof_uvarint(s) + s
