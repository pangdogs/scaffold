extends RefCounted

const Fnv32a = preload("fnv32a.gd")

const TARGET_SERVICE := 0x53
const TARGET_RUNTIME := 0x52
const TARGET_ENTITY := 0x45
const TARGET_CLIENT := 0x43

const FLAG_SHORT := 1
const FLAG_EXCLUDE_SRC := 2

class CallPath:
	extends RefCounted

	var target_kind: int
	var id: String
	var scr: String
	var method: String

class Cached:
	extends RefCounted

	var scr: String
	var method: String

	func _init(script: String, method: String):
		self.scr = script
		self.method = method

class Cache:
	var _cache: Dictionary[int, Cached] = {}

	func intern(script: String, method: String) -> void:
		var idx := reduce(script, method)
		var existing: Cached = _cache.get_or_add(idx, Cached.new(script, method))
		if existing.scr != script or existing.method != method:
			push_error("intern call path cached index %d conflict: existing %s.%s vs new %s.%s; rename the script or method to change the generated call path id" % [idx, existing.scr, existing.method, script, method])

	func intern_script(script: String, target: GDScript) -> void:
		if not is_instance_valid(target):
			return
		for method_info in target.get_script_method_list():
			var method_name: String = method_info["name"]
			if method_name.begins_with("_"):
				continue
			var flags := int(method_info.get("flags", 0))
			if (flags & METHOD_FLAG_STATIC) != 0:
				continue
			intern(script, method_name)

	func intern_object(script: String, target: Object) -> void:
		if not is_instance_valid(target):
			return
		var target_script := target.get_script() as GDScript
		if target_script == null:
			return
		intern_script(script, target_script)

	func clear() -> void:
		_cache.clear()

	static func reduce(script: String, method: String) -> int:
		var hasher := Fnv32a.new()
		hasher.write_bytes(script.to_utf8_buffer())
		hasher.write_byte(0)
		hasher.write_bytes(method.to_utf8_buffer())
		return hasher.sum32()

	func inflate(idx: int) -> Cached:
		return _cache.get(idx)

static var cache := Cache.new()

static func encode(target_kind: int, script: String, method: String, id: String = "", short: bool = true) -> PackedByteArray:
	var buff := StreamPeerBuffer.new()
	buff.big_endian = false

	buff.put_u8(target_kind)

	var flags := 0
	if short:
		flags |= FLAG_SHORT
	buff.put_u8(flags)

	match target_kind:
		TARGET_SERVICE, TARGET_CLIENT:
			pass
		TARGET_RUNTIME, TARGET_ENTITY:
			buff.put_data(id.to_utf8_buffer())
			buff.put_u8(0)
		_:
			return PackedByteArray()

	if short:
		buff.put_u32(cache.reduce(script, method))
	else:
		buff.put_data(script.to_utf8_buffer())
		buff.put_u8(0)
		buff.put_data(method.to_utf8_buffer())
		buff.put_u8(0)

	return buff.data_array

static func parse(data: PackedByteArray) -> CallPath:
	if data.size() < 2:
		return null

	var buff := StreamPeerBuffer.new()
	buff.big_endian = false
	buff.data_array = data

	var cp := CallPath.new()
	if not _has_remaining(buff, 1):
		return null
	cp.target_kind = buff.get_u8()

	if not _has_remaining(buff, 1):
		return null
	var flags := buff.get_u8()
	var short := (flags & FLAG_SHORT) != 0

	match cp.target_kind:
		TARGET_SERVICE, TARGET_CLIENT:
			pass
		TARGET_RUNTIME, TARGET_ENTITY:
			if not _has_remaining(buff, 1):
				return null
			cp.id = _read_c_string(buff)
		_:
			return null

	if short:
		if not _has_remaining(buff, 4):
			return null
		var cached := cache.inflate(buff.get_u32())
		if cached == null:
			return null
		cp.scr = cached.scr
		cp.method = cached.method
	else:
		if not _has_remaining(buff, 1):
			return null
		cp.scr = _read_c_string(buff)
		if not _has_remaining(buff, 1):
			return null
		cp.method = _read_c_string(buff)

	return cp

static func _has_remaining(buff: StreamPeerBuffer, size: int) -> bool:
	return buff.get_position() + size <= buff.data_array.size()

static func _read_c_string(buff: StreamPeerBuffer) -> String:
	var data := buff.data_array
	var start := buff.get_position()

	for i in range(start, data.size()):
		if data[i] == 0:
			buff.seek(i + 1)
			return data.slice(start, i).get_string_from_utf8()

	return ""
