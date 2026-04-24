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
# Shared helpers used by generated Excel table access code.
# The generated code relies on this runtime for index conversion,
# dictionary key ordering, and a few small binary-search utilities.
class_name ExcelUtils
extends RefCounted

# Runtime helper for chunked Excel tables.
# It owns per-chunk cache and coordinates once-only background loads.
class ChunkLoader:
	extends RefCounted

	# Per-chunk loading state shared by concurrent callers.
	class ChunkState:
		extends RefCounted

		const READY: int = -1
		const LOADING: int = -2
		const LOADED: int = -3

		var mutex := Mutex.new()
		var rows: Array = []
		var task_id := READY

	var _chunk_base_path := ""
	var _chunk_states: Array[ChunkState] = []
	var _message_factory: Callable

	# Pre-allocates state objects for all chunks described by the manifest.
	func _init(chunk_base_path: String, chunk_count: int, message_factory: Callable) -> void:
		_chunk_base_path = chunk_base_path
		_chunk_states.resize(chunk_count)
		for chunk_index in range(chunk_count):
			_chunk_states[chunk_index] = ChunkState.new()
		_message_factory = message_factory

	func rows(chunk_index: int) -> Array:
		if chunk_index < 0 or chunk_index >= _chunk_states.size():
			return []
		var state := _chunk_states[chunk_index]
		state.mutex.lock()
		var chunk_rows := state.rows
		state.mutex.unlock()
		return chunk_rows

	# Synchronously ensures one chunk is loaded and cached.
	func ensure_loaded(chunk_index: int) -> bool:
		if chunk_index < 0 or chunk_index >= _chunk_states.size():
			return false
		var state := _chunk_states[chunk_index]
		var task_id := _ensure_task_started(state, chunk_index)
		match task_id:
			ChunkState.LOADING:
				state.mutex.lock()
				task_id = state.task_id
				state.mutex.unlock()
				if task_id == ChunkState.LOADED:
					return true
				while !WorkerThreadPool.is_task_completed(task_id):
					OS.delay_usec(100)
			ChunkState.LOADED:
				return true
			_:
				WorkerThreadPool.wait_for_task_completion(task_id)
		return true

	# Asynchronously ensures one chunk is loaded.
	# The first caller starts the task; later callers await the same task completion.
	func ensure_loaded_async(chunk_index: int) -> bool:
		if chunk_index < 0 or chunk_index >= _chunk_states.size():
			return false
		var state := _chunk_states[chunk_index]
		var task_id := _ensure_task_started(state, chunk_index)
		match task_id:
			ChunkState.LOADING:
				state.mutex.lock()
				task_id = state.task_id
				state.mutex.unlock()
				if task_id == ChunkState.LOADED:
					return true
				if Thread.is_main_thread():
					var tree := Engine.get_main_loop() as SceneTree
					while !WorkerThreadPool.is_task_completed(task_id):
						await tree.process_frame
				else:
					while !WorkerThreadPool.is_task_completed(task_id):
						OS.delay_usec(100)
			ChunkState.LOADED:
				return true
			_:
				WorkerThreadPool.wait_for_task_completion(task_id)
		return true

	# Starts the chunk load task at most once and returns the shared task id.
	func _ensure_task_started(state: ChunkState, chunk_index: int) -> int:
		state.mutex.lock()
		var task_id: int
		match state.task_id:
			ChunkState.READY:
				state.task_id = WorkerThreadPool.add_task(_load_rows_task.bind(state, chunk_index))
				task_id = state.task_id
			ChunkState.LOADED:
				task_id = ChunkState.LOADED
			_:
				task_id = ChunkState.LOADING
		state.mutex.unlock()
		return task_id

	# Worker entry point. It deserializes one chunk file into row data.
	func _load_rows_task(state: ChunkState, chunk_index: int) -> void:
		var chunk_rows: Array = _load_rows(_chunk_path(chunk_index))
		state.mutex.lock()
		state.rows = chunk_rows
		state.task_id = ChunkState.LOADED
		state.mutex.unlock()

	# Builds chunk file path.
	func _chunk_path(chunk_index: int) -> String:
		return "%s.chk_%d" % [_chunk_base_path, chunk_index]

	# Builds a fresh message instance and reads chunk rows from file.
	func _load_rows(path: String) -> Array:
		var start_usec := Time.get_ticks_usec()
		var msg = _message_factory.call()
		var file := FileAccess.open(path, FileAccess.READ)
		if file == null:
			return []
		var stream := ProtoInputFile.new(file)
		if !msg.deserialize(stream):
			return []
		var elapsed_ms := float(Time.get_ticks_usec() - start_usec) / 1000.0
		print("excel table chunk file loaded, file_path=%s, rows=%d, elapsed_ms=%.3f" % [path, msg.Rows.size(), elapsed_ms])
		return msg.Rows

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

# Performs a binary search over ordered unsigned 64-bit values and returns the matching position.
static func binary_search_u64(items: Array[int], value: int) -> int:
	var low := 0
	var high := items.size() - 1
	while low <= high:
		@warning_ignore("integer_division")
		var mid := (low + high) / 2
		var cmp := _compare_u64(items[mid], value)
		if cmp < 0:
			low = mid + 1
		elif cmp > 0:
			high = mid - 1
		else:
			return mid
	return -1

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
