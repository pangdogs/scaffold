class_name GolaxyRPCLI
extends Node

const CallPath = preload("call_path.gd")
const ByteStream = preload("byte_stream.gd")

class ScriptInfo:
	extends RefCounted

	var target_id: int = 0
	var target: Object = null
	var auto_unbind: Callable = Callable()
	var methods: Dictionary = {}
	var priority: int = 0
	var version: int = 0

	func _init(target: Object, priority: int, version: int) -> void:
		if is_instance_valid(target):
			self.target_id = target.get_instance_id()
			self.target = target
		self.methods = _extract_methods(target)
		self.priority = priority
		self.version = version

	func _notification(what: int) -> void:
		if what == NOTIFICATION_PREDELETE:
			var target_node := target as Node
			if target_node != null:
				if auto_unbind.is_valid() and target_node.tree_exited.is_connected(auto_unbind):
					target_node.tree_exited.disconnect(auto_unbind)
			target = null
			auto_unbind = Callable()
			methods.clear()

	func contains_method(name: String) -> bool:
		return methods.has(name)

	func method_info(name: String) -> Dictionary:
		if not contains_method(name):
			return {}
		return methods[name]

	static func _sort_less(a: ScriptInfo, b: ScriptInfo) -> bool:
		if a.priority == b.priority:
			return a.version < b.version
		return a.priority < b.priority

	static func _extract_methods(target: Object) -> Dictionary:
		var ret := {}
		if not is_instance_valid(target):
			return ret
		var target_script := target.get_script() as GDScript
		if target_script == null:
			return ret
		for method_info in target_script.get_script_method_list():
			var method_name: String = method_info["name"]
			if method_name.begins_with("_"):
				continue
			var flags := int(method_info.get("flags", 0))
			if (flags & METHOD_FLAG_STATIC) != 0:
				continue
			ret[method_name] = method_info
		return ret

class Result:
	extends RefCounted

	var value: Variant = null
	var error: GAPVariants.GAPError = null

	func _init(value: Variant, error: GAPVariants.GAPError = null):
		self.value = value
		self.error = error

	func ok() -> bool:
		return error == null

	func _to_string() -> String:
		return "Result{value=%s, error=%s}" % [value, error]

var _connecting: bool = false
var _client: GolaxyClient = null
var _scripts: Dictionary[String, Array] = {}
var _scripts_version := 0
var _remote_clock: GolaxyClient.TimeSample = null
var _logger := GolaxyLogger.new("GolaxyRPCLI", get_instance_id(), true)

var logger: GolaxyLogger:
	get:
		return _logger

func _exit_tree() -> void:
	close(ERR_UNAVAILABLE, "game exiting")
	unbind_all()

func connect_to_async(
	endpoint: String,
	protocol: int = GolaxyClient.PROTOCOL_TCP,
	user_id: String = "",
	token: String = "",
	timeout_ms: int = GolaxyClient.DEFAULT_CONNECT_TIMEOUT_MS,
	reconnect_max_attempts: int = GolaxyClient.DEFAULT_RECONNECT_MAX_ATTEMPTS,
	reconnect_delay_ms: int = GolaxyClient.DEFAULT_RECONNECT_DELAY_MS,
	reconnect_timeout_ms: int = GolaxyClient.DEFAULT_RECONNECT_TIMEOUT_MS,
	heartbeat_interval_ms: int = GolaxyClient.DEFAULT_HEARTBEAT_INTERVAL_MS,
	heartbeat_timeout_ms: int = GolaxyClient.DEFAULT_HEARTBEAT_TIMEOUT_MS,
	sample_remote_clock_times: int = 3
) -> bool:
	if _connecting:
		return false
	_connecting = true
	_logger.debug("connect begin endpoint=%s protocol=%d user_id=%s timeout_ms=%d reconnect_max=%d reconnect_delay_ms=%d", [
		endpoint,
		protocol,
		user_id,
		timeout_ms,
		reconnect_max_attempts,
		reconnect_delay_ms,
	])

	var client := GolaxyClient.new(_logger)
	client.data_received.connect(_handle_data, CONNECT_APPEND_SOURCE_OBJECT)

	var poll_timer := Timer.new()
	poll_timer.wait_time = 0.001
	poll_timer.ignore_time_scale = true
	poll_timer.process_callback = Timer.TIMER_PROCESS_IDLE
	poll_timer.process_mode = Node.PROCESS_MODE_ALWAYS
	poll_timer.timeout.connect(func():
		client.poll()
		if client.is_closed():
			client.data_received.disconnect(_handle_data)
			poll_timer.queue_free()
	)
	add_child(poll_timer)
	poll_timer.start()

	if _client != null:
		_client.close(ERR_UNAVAILABLE, "client replaced")
	_client = client
	_remote_clock = null

	var ok := await _client.connect_to_async(
		endpoint,
		protocol,
		user_id,
		token,
		timeout_ms,
		reconnect_max_attempts,
		reconnect_delay_ms,
		reconnect_timeout_ms,
		heartbeat_interval_ms,
		heartbeat_timeout_ms
	)
	if !ok:
		_client.close(ERR_CANT_CONNECT, "connecting failed")
		_connecting = false
		_logger.debug("connect failed endpoint=%s", [endpoint])
		return false
	_remote_clock = await _sample_best_remote_clock_async(_client, sample_remote_clock_times)

	_connecting = false
	_logger.debug("connect ok endpoint=%s session=%s", [endpoint, _client.session_id()])
	return true

func close(error: int = OK, error_message: String = "") -> void:
	_logger.debug("close error=%d message=%s", [error, error_message])
	if _client != null:
		_client.close(error, error_message)

func client() -> GolaxyClient:
	return _client

func remote_clock() -> GolaxyClient.TimeSample:
	return _remote_clock

func bind(script: String, target: Object, priority: int = 0) -> bool:
	if not is_instance_valid(target):
		return false
	var script_infos := _scripts.get(script, []) as Array
	if not script_infos.is_empty():
		script_infos = script_infos.duplicate()
	_scripts_version += 1
	for script_info: ScriptInfo in script_infos:
		if script_info.target_id == target.get_instance_id():
			var old_priority := script_info.priority
			script_info.priority = priority
			script_infos.sort_custom(ScriptInfo._sort_less)
			_scripts[script] = script_infos
			_logger.debug("script target binding updated, script=%s, target_id=%d, priority=(%d->%d)", [script, script_info.target_id, old_priority, priority])
			return true
	var script_info := ScriptInfo.new(target, priority, _scripts_version)
	var target_node := target as Node
	if target_node != null:
		var script_info_id := script_info.get_instance_id()
		script_info.auto_unbind = func(): _auto_unbind(script, script_info_id)
		target_node.tree_exited.connect(script_info.auto_unbind, CONNECT_ONE_SHOT)
	script_infos.append(script_info)
	script_infos.sort_custom(ScriptInfo._sort_less)
	_scripts[script] = script_infos
	_logger.debug("script target bound, script=%s, target_id=%d, priority=%d, methods=%s", [script, script_info.target_id, priority, script_info.methods.keys()])
	CallPath.cache.intern_object(script, target)
	return true

func unbind(script: String, target_id: int) -> void:
	var script_infos := _scripts.get(script, []) as Array
	if script_infos.is_empty():
		return
	for i in range(script_infos.size()):
		var script_info := script_infos[i] as ScriptInfo
		if script_info.target_id == target_id:
			script_infos = script_infos.duplicate()
			script_infos.remove_at(i)
			_scripts_version += 1
			if script_infos.is_empty():
				_scripts.erase(script)
			else:
				_scripts[script] = script_infos
			_logger.debug("script target unbound, script=%s, target_id=%d", [script, target_id])
			return

func unbind_script(script: String) -> void:
	var script_infos := _scripts.get(script, []) as Array
	var target_ids := script_infos.map(func(script_info: ScriptInfo): return script_info.target_id)
	_scripts.erase(script)
	_logger.debug("script targets unbound, script=%s, target_ids=%s", [script, target_ids])

func unbind_all() -> void:
	var bindings := {}
	for script: String in _scripts:
		var script_infos := _scripts[script] as Array
		bindings[script] = script_infos.map(func(script_info: ScriptInfo): return script_info.target_id)
	_scripts.clear()
	_logger.debug("all script targets unbound, bindings=%s", [bindings])

func rpc_async(
	service: String,
	component: String,
	method: String,
	args: Array = [],
	timeout_ms: int = GolaxyClient.DEFAULT_FUTURE_TIMEOUT_MS,
	reduce_call_path: bool = true
) -> Result:
	if _client == null:
		var error := GAPVariants.GAPError.new(-1, "client is not connected")
		_logger.error("failed to rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return Result.new(null, error)

	var future := _client.new_future(timeout_ms)
	if future == null:
		var error := GAPVariants.GAPError.new(-1, "client is not established")
		_logger.error("failed to rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return Result.new(null, error)

	var rpc_args := GAPVariants.GAPArray.from_native(args)
	if rpc_args == null:
		var error := GAPVariants.GAPError.new(-1, "encode rpc args failed")
		_logger.error("failed to rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		_cancel_future(future, ERR_INVALID_PARAMETER, error.to_string())
		return Result.new(null, error)

	var rpc_path := CallPath.encode(CallPath.TARGET_ENTITY, component, method, "", reduce_call_path)
	if rpc_path.is_empty():
		var error := GAPVariants.GAPError.new(-1, "encode rpc call path failed")
		_logger.error("failed to rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		_cancel_future(future, ERR_INVALID_PARAMETER, error.to_string())
		return Result.new(null, error)

	var msg := GAPMessages.MsgRPCRequest.new()
	msg.corr_id = future.corr_id
	msg.path = rpc_path
	msg.args = rpc_args

	var msg_buf := _encode_message(msg)
	if msg_buf.is_empty():
		var error := GAPVariants.GAPError.new(-1, "encode rpc request failed")
		_logger.error("failed to rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		_cancel_future(future, ERR_INVALID_PARAMETER, error.to_string())
		return Result.new(null, error)

	var forward_msg := GAPMessages.MsgForward.new()
	forward_msg.dst = service
	forward_msg.corr_id = msg.corr_id
	forward_msg.trans_id = msg.msg_id()
	forward_msg.trans_data = msg_buf

	var packet := GAPCodec.encode_packet(forward_msg)
	if packet.is_empty():
		var error := GAPVariants.GAPError.new(-1, "encode rpc packet failed")
		_logger.error("failed to rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		_cancel_future(future, ERR_INVALID_PARAMETER, error.to_string())
		return Result.new(null, error)

	if !_client.send_data(packet):
		var error := GAPVariants.GAPError.new(-1, "send rpc packet failed")
		_logger.error("failed to rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		_cancel_future(future, ERR_CONNECTION_ERROR, error.to_string())
		return Result.new(null, error)

	var result: Array = await future.completed
	if result[1] != OK:
		var error := GAPVariants.GAPError.new(-1, result[2])
		_logger.error("failed to rpc, service=%s, component=%s, method=%s, args=%s, corr_id=%d, error=%s", [service, component, method, args, future.corr_id, error])
		return Result.new(null, error)

	if _logger.debug_enabled:
		_logger.debug("rpc completed, service=%s, component=%s, method=%s, args=%s, corr_id=%d, result=%s", [service, component, method, args, future.corr_id, result[0]])
	return result[0] as Result

func oneway_rpc(
	service: String,
	component: String,
	method: String,
	args: Array = [],
	reduce_call_path: bool = true
) -> bool:
	if _client == null:
		var error := GAPVariants.GAPError.new(-1, "client is not connected")
		_logger.error("failed to oneway rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return false

	if not _client.is_established():
		var error := GAPVariants.GAPError.new(-1, "client is not established")
		_logger.error("failed to oneway rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return false

	var rpc_args := GAPVariants.GAPArray.from_native(args)
	if rpc_args == null:
		var error := GAPVariants.GAPError.new(-1, "encode rpc args failed")
		_logger.error("failed to oneway rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return false

	var rpc_path := CallPath.encode(CallPath.TARGET_ENTITY, component, method, "", reduce_call_path)
	if rpc_path.is_empty():
		var error := GAPVariants.GAPError.new(-1, "encode rpc call path failed")
		_logger.error("failed to oneway rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return false

	var msg := GAPMessages.MsgOnewayRPC.new()
	msg.path = rpc_path
	msg.args = rpc_args

	var msg_buf := _encode_message(msg)
	if msg_buf.is_empty():
		var error := GAPVariants.GAPError.new(-1, "encode rpc request failed")
		_logger.error("failed to oneway rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return false

	var forward_msg := GAPMessages.MsgForward.new()
	forward_msg.dst = service
	forward_msg.trans_id = msg.msg_id()
	forward_msg.trans_data = msg_buf

	var packet := GAPCodec.encode_packet(forward_msg)
	if packet.is_empty():
		var error := GAPVariants.GAPError.new(-1, "encode rpc packet failed")
		_logger.error("failed to oneway rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return false

	if not _client.send_data(packet):
		var error := GAPVariants.GAPError.new(-1, "send rpc packet failed")
		_logger.error("failed to oneway rpc, service=%s, component=%s, method=%s, args=%s, error=%s", [service, component, method, args, error])
		return false

	if _logger.debug_enabled:
		_logger.debug("oneway rpc completed, service=%s, component=%s, method=%s, args=%s", [service, component, method, args])
	return true

func _handle_data(data: PackedByteArray, client: GolaxyClient) -> void:
	var packet := GAPCodec.decode_packet(data)
	if packet == null:
		_logger.error("decode GAP packet failed")
		return

	match packet.head.msg_id:
		GAPMessages.MSG_ID_ONEWAY_RPC:
			var msg := packet.body as GAPMessages.MsgOnewayRPC
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgOnewayRPC")
			await _accept_notify_async(msg)
		GAPMessages.MSG_ID_RPC_REQUEST:
			var msg := packet.body as GAPMessages.MsgRPCRequest
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgRPCRequest")
			await _accept_request_async(packet.head.src, msg)
		GAPMessages.MSG_ID_RPC_REPLY:
			var msg := packet.body as GAPMessages.MsgRPCReply
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgRPCReply")
			_resolve_reply(client, msg)
		_:
			_logger.warning("unexpected GAP packet: %d", [packet.head.msg_id])

func _accept_notify_async(msg: GAPMessages.MsgOnewayRPC) -> void:
	var rpc_path := CallPath.parse(msg.path)
	if rpc_path == null:
		_logger.error("parse call path failed")
		return
	if rpc_path.target_kind != CallPath.TARGET_CLIENT:
		_logger.error("unexpected call path target_kind: %d", [rpc_path.target_kind])
		return

	var result := await _call_script_async(rpc_path.scr, rpc_path.method, msg.args, false)
	if result != null:
		if not result.ok():
			_logger.error("failed to call script, script=%s, method=%s, args=%s, result=%s", [rpc_path.scr, rpc_path.method, msg.args, result])
		else:
			if _logger.debug_enabled:
				_logger.debug("script called, script=%s, method=%s, args=%s", [rpc_path.scr, rpc_path.method, msg.args])

func _accept_request_async(src: GAPMessages.Origin, msg: GAPMessages.MsgRPCRequest) -> void:
	var rpc_path := CallPath.parse(msg.path)
	if rpc_path == null:
		_logger.error("parse call path failed, corr_id=%d", [msg.corr_id])
		_send_reply(src, msg.corr_id, Result.new(null, GAPVariants.GAPError.new(-1, "invalid call path")))
		return
	if rpc_path.target_kind != CallPath.TARGET_CLIENT:
		_logger.error("unexpected call path target_kind: %d, corr_id=%d", [rpc_path.target_kind, msg.corr_id])
		_send_reply(src, msg.corr_id, Result.new(null, GAPVariants.GAPError.new(-1, "invalid call path target_kind")))
		return

	var result := await _call_script_async(rpc_path.scr, rpc_path.method, msg.args, true)
	if not result.ok():
		_logger.error("failed to call script, script=%s, method=%s, args=%s, corr_id=%d, result=%s", [rpc_path.scr, rpc_path.method, msg.args, msg.corr_id, result])
	else:
		if _logger.debug_enabled:
			_logger.debug("script called, script=%s, method=%s, args=%s, corr_id=%d, result=%s", [rpc_path.scr, rpc_path.method, msg.args, msg.corr_id, result])

	_send_reply(src, msg.corr_id, result)

func _resolve_reply(client: GolaxyClient, msg: GAPMessages.MsgRPCReply) -> void:
	var future := client.get_future(msg.corr_id)
	if future == null:
		_logger.warning("unexpected resolve rpc reply corr_id: %d", [msg.corr_id])
		return

	if msg.error.code != 0:
		_logger.error("rpc reply resolved, corr_id=%d, error=%s", [msg.corr_id, msg.error])
		future.resolve(Result.new(null, msg.error))
		return

	var rets := msg.rets.to_native()
	var ret_value: Variant = rets
	if rets.size() <= 0:
		ret_value = null
	elif rets.size() == 1:
		ret_value = rets[0]

	if _logger.debug_enabled:
		_logger.debug("rpc reply resolved, corr_id=%d, result=%s", [msg.corr_id, ret_value])
	future.resolve(Result.new(ret_value))

func _send_reply(src: GAPMessages.Origin, corr_id: int, result: Result) -> void:
	if corr_id == 0:
		return

	var msg := GAPMessages.MsgRPCReply.new()
	msg.corr_id = corr_id

	if result.ok():
		if result.value != null:
			var ret := GAPVariants.to_variant(result.value)
			if ret != null:
				if ret.type_id() == GAPVariants.TYPE_ID_ARRAY:
					var arr := ret.value() as GAPVariants.GAPArray
					msg.rets.items.append_array(arr.items)
				else:
					msg.rets.items.append(ret)
	else:
		msg.error = result.error

	var msg_buf := _encode_message(msg)
	if msg_buf.is_empty():
		_logger.error("encode rpc reply failed")
		return

	var forward_msg := GAPMessages.MsgForward.new()
	forward_msg.dst = src.addr
	forward_msg.corr_id = corr_id
	forward_msg.trans_id = msg.msg_id()
	forward_msg.trans_data = msg_buf

	var packet := GAPCodec.encode_packet(forward_msg)
	if packet.is_empty():
		_logger.error("encode rpc reply packet failed")
		return

	if !_client.send_data(packet):
		_logger.error("send rpc reply packet failed")
		return

	if _logger.debug_enabled:
		_logger.debug("rpc replied, corr_id=%d, result=%s", [corr_id, result])

func _call_script_async(script: String, method: String, args: GAPVariants.GAPArray, need_result: bool) -> Result:
	var has_result := _logger.debug_enabled or need_result

	var script_infos := _scripts.get(script, []) as Array
	if script_infos.is_empty():
		return Result.new(null, GAPVariants.GAPError.new(-1, "script not found")) if has_result else null

	var call_args := args.to_native()
	var call_times := 0
	var result_captured := not need_result
	var result_value: Variant = null

	for script_info: ScriptInfo in script_infos:
		if not is_instance_valid(script_info.target):
			continue
		if not script_info.contains_method(method):
			continue
		var call_ret: Variant = await script_info.target.callv(method, call_args)
		if _logger.debug_enabled:
			_logger.debug("script method called, target_id=%d, script=%s, method=%s, call_args=%s, call_ret=%s", [script_info.target_id, script, method, call_args, call_ret])
		if call_ret != null:
			if not result_captured:
				result_value = call_ret
				result_captured = true
		call_times += 1

	if call_times <= 0:
		return Result.new(null, GAPVariants.GAPError.new(-1, "method not found")) if has_result else null

	return Result.new(result_value) if has_result else null

func _auto_unbind(script: String, script_info_id: int) -> void:
	var script_infos := _scripts.get(script, []) as Array
	if script_infos.is_empty():
		return
	var idx := script_infos.find_custom(func(script_info: ScriptInfo): return script_info.get_instance_id() == script_info_id)
	if idx < 0:
		_logger.debug("script auto unbind skipped, script=%s, script_info_id=%d", [script, script_info_id])
		return
	script_infos = script_infos.duplicate()
	var script_info := script_infos[idx] as ScriptInfo
	script_infos.remove_at(idx)
	_scripts_version += 1
	if script_infos.is_empty():
		_scripts.erase(script)
	else:
		_scripts[script] = script_infos
	_logger.debug("script auto unbound, script=%s, target_id=%d, script_info_id=%d", [script, script_info.target_id, script_info_id])

static func _encode_message(msg: GAPMessages.Msg) -> PackedByteArray:
	if msg == null:
		return PackedByteArray()
	var stream := ByteStream.new()
	if !msg.serialize(stream):
		return PackedByteArray()
	return stream.bytes()

static func _cancel_future(future: GolaxyClient.Future, error: int, message: String) -> void:
	if future == null or future.is_done():
		return
	future.reject(error, message)

static func _sample_best_remote_clock_async(client: GolaxyClient, samples: int = 3) -> GolaxyClient.TimeSample:
	var best: GolaxyClient.TimeSample = null
	for i in range(maxi(1, samples)):
		var time_sample := await client.probe_time_async()
		if time_sample == null:
			continue
		if best == null or time_sample.rtt() < best.rtt():
			best = time_sample
	return best
