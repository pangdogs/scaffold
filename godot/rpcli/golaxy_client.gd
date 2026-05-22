class_name GolaxyClient
extends RefCounted

signal connecting()
signal connect_completed(error: int, error_message: String)
signal connected()
signal reconnecting(attempt: int, max_attempts: int)
signal reconnect_completed(attempt: int, max_attempts: int, error: int, error_message: String)
signal reconnected(attempt: int, max_attempts: int)
signal data_received(data: PackedByteArray)
signal disconnected()

const PHASE_IDLE := 0
const PHASE_CONNECTING := 1
const PHASE_WAIT_HELLO := 2
const PHASE_WAIT_FINISHED := 3
const PHASE_CONNECTED := 4
const PHASE_CLOSED := 5

const RECONNECT_IDLE := 0
const RECONNECT_WAITING := 1
const RECONNECT_CONNECTING := 2

const PROTOCOL_TCP := 0
const PROTOCOL_WEBSOCKET := 1

const DEFAULT_CONNECT_TIMEOUT_MS := 7000
const DEFAULT_FUTURE_TIMEOUT_MS := 15000
const DEFAULT_RECONNECT_MAX_ATTEMPTS := 3
const DEFAULT_RECONNECT_DELAY_MS := 1000
const DEFAULT_RECONNECT_TIMEOUT_MS := 4000
const DEFAULT_HEARTBEAT_INTERVAL_MS := 3000
const DEFAULT_HEARTBEAT_TIMEOUT_MS := 10000

class Frame:
	extends RefCounted

	var seq: int = 0
	var sent_ms: int = 0
	var offset: int = 0
	var data: PackedByteArray = PackedByteArray()

	func _init(seq: int = 0, data: PackedByteArray = PackedByteArray()) -> void:
		self.seq = seq
		self.data = data

class Future:
	extends RefCounted

	signal completed(value: Variant, error: int, error_message: String)

	var corr_id: int = 0
	var deadline_ms: int = 0
	var _done := false

	func _init(corr_id: int = 0, deadline_ms: int = 0) -> void:
		self.corr_id = corr_id
		self.deadline_ms = deadline_ms

	func is_done() -> bool:
		return _done

	func resolve(value: Variant = null) -> void:
		if _done:
			return
		_done = true
		completed.emit(value, OK, "")

	func reject(error: int = FAILED, error_message: String = "") -> void:
		if _done:
			return
		_done = true
		completed.emit(null, error if error != OK else FAILED, error_message)

class TimeSample:
	extends RefCounted

	var origin_ms: int = 0
	var receive_ms: int = 0
	var transmit_ms: int = 0
	var destination_ms: int = 0
	var remote_zone_offset: int = 0

	func rtt() -> int:
		return (destination_ms - origin_ms) - (transmit_ms - receive_ms)

	func offset() -> int:
		return int(((receive_ms - origin_ms) + (transmit_ms - destination_ms)) / 2.0)

	func remote_time() -> int:
		return destination_ms + offset()

	func remote_now() -> int:
		return GolaxyClient._local_unix_time_ms() + offset()

var _conn: GolaxyConn = null
var _phase: int = PHASE_IDLE
var _endpoint: String = ""
var _protocol: int = PROTOCOL_TCP
var _user_id: String = ""
var _token: String = ""
var _session_id: String = ""
var _send_window: Array[Frame] = []
var _send_seq: int = 0
var _recv_seq: int = 0
var _next_corr_id: int = randi()
var _recv_buff: PackedByteArray = PackedByteArray()
var _futures: Dictionary[int, Future] = {}
var _expect_auth := false
var _connect_deadline_ms: int = 0
var _reconnect_max_attempts: int = DEFAULT_RECONNECT_MAX_ATTEMPTS
var _reconnect_delay_ms: int = DEFAULT_RECONNECT_DELAY_MS
var _reconnect_timeout_ms: int = DEFAULT_RECONNECT_TIMEOUT_MS
var _reconnect_state: int = RECONNECT_IDLE
var _reconnect_attempts: int = 0
var _next_reconnect_ms: int = 0
var _reconnect_error: int = OK
var _reconnect_error_message: String = ""
var _heartbeat_interval_ms: int = DEFAULT_HEARTBEAT_INTERVAL_MS
var _heartbeat_timeout_ms: int = DEFAULT_HEARTBEAT_TIMEOUT_MS
var _last_recv_ms: int = 0
var _last_ping_ms: int = 0
var _close_error: int = OK
var _close_error_message: String = ""
var _logger := GolaxyLogger.new("GolaxyClient", get_instance_id())

var logger: GolaxyLogger:
	get:
		return _logger

#region Lifecycle

func _init(logger: GolaxyLogger = null) -> void:
	if logger != null:
		_logger = logger.named("GolaxyClient", get_instance_id())

#endregion

#region Public API

func connect_to_async(
	endpoint: String,
	protocol: int = PROTOCOL_TCP,
	user_id: String = "",
	token: String = "",
	timeout_ms: int = DEFAULT_CONNECT_TIMEOUT_MS,
	reconnect_max_attempts: int = DEFAULT_RECONNECT_MAX_ATTEMPTS,
	reconnect_delay_ms: int = DEFAULT_RECONNECT_DELAY_MS,
	reconnect_timeout_ms: int = DEFAULT_RECONNECT_TIMEOUT_MS,
	heartbeat_interval_ms: int = DEFAULT_HEARTBEAT_INTERVAL_MS,
	heartbeat_timeout_ms: int = DEFAULT_HEARTBEAT_TIMEOUT_MS
) -> bool:
	if _phase != PHASE_IDLE:
		return false
	_endpoint = endpoint
	_protocol = protocol
	_user_id = user_id
	_token = token
	_reconnect_max_attempts = maxi(reconnect_max_attempts, 0)
	_reconnect_delay_ms = maxi(reconnect_delay_ms, 0)
	_reconnect_timeout_ms = maxi(reconnect_timeout_ms, 0)
	_heartbeat_interval_ms = maxi(heartbeat_interval_ms, 0)
	_heartbeat_timeout_ms = maxi(heartbeat_timeout_ms, 0)
	if not _begin_connect(timeout_ms):
		return false
	var result: Array = await connect_completed
	return result[0] == OK

func close(error: int = OK, error_message: String = "") -> void:
	if _phase == PHASE_CLOSED:
		return

	_logger.debug("close error=%d message=%s", [error, error_message])

	var should_emit_connect_completed := _is_connecting() and not _is_reconnecting()
	var should_emit_disconnected := _phase == PHASE_CONNECTED or _is_reconnecting()

	_reset_for_close(error, error_message)

	if _conn != null:
		_conn.close()
		_conn = null

	for corr_id in _futures.keys():
		var future := _futures[corr_id]
		if not future.is_done():
			future.reject(error, error_message)
	_futures.clear()

	if should_emit_connect_completed:
		connect_completed.emit(error, error_message)
	if should_emit_disconnected:
		disconnected.emit()

func poll() -> void:
	if _phase == PHASE_CLOSED:
		return
	if _conn != null:
		_conn.poll()
	_process_connect_timeout()
	_process_io()
	_process_phase()
	_process_heartbeat()
	_process_reconnect()
	_process_futures()

func send_data(data: PackedByteArray) -> bool:
	if not is_established():
		_logger.error("failed to send data, client is not established")
		return false
	var msg := GTPMessages.MsgPayload.new()
	msg.data = data
	if not _send_reliable_message(msg):
		_logger.error("failed to send data, send GTP message failed")
		return false
	return true

func probe_time_async(timeout_ms: int = DEFAULT_FUTURE_TIMEOUT_MS) -> TimeSample:
	var future := _send_probe_time(timeout_ms)
	if future == null:
		_logger.error("failed to probe time, client is not established")
		return null
	var result: Array = await future.completed
	if result[1] != OK:
		_logger.error("failed to probe time, error=%d, error_message=%s", [result[1], result[2]])
		return null
	return result[0]

func new_future(timeout_ms: int = DEFAULT_FUTURE_TIMEOUT_MS) -> Future:
	if not is_established():
		_logger.error("failed to new future, client is not established")
		return null
	var corr_id := _next_corr_id
	_next_corr_id += 1
	if _next_corr_id == 0:
		_next_corr_id += 1
	var deadline_ms := _local_unix_time_ms() + maxi(timeout_ms, 0)
	var future := Future.new(corr_id, deadline_ms)
	_futures[corr_id] = future
	return future

func get_future(corr_id: int) -> Future:
	var future := _futures.get(corr_id)
	if future == null or future.is_done():
		return null
	return future

func user_id() -> String:
	return _user_id

func token() -> String:
	return _token

func session_id() -> String:
	return _session_id

func endpoint() -> String:
	return _endpoint

func is_connecting() -> bool:
	return _is_connecting() and not _is_reconnecting()

func is_established() -> bool:
	return _phase == PHASE_CONNECTED or _is_reconnecting()

func is_closed() -> bool:
	return _phase == PHASE_CLOSED

func close_error() -> int:
	return _close_error

func close_error_message() -> String:
	return _close_error_message

#endregion

#region Connection Lifecycle

func _begin_connect(timeout_ms: int) -> bool:
	_reset_for_initial_connect(timeout_ms)
	_logger.debug("connect begin endpoint=%s protocol=%d timeout_ms=%d reconnect_max=%d reconnect_delay_ms=%d reconnect_timeout_ms=%d heartbeat_interval_ms=%d heartbeat_timeout_ms=%d", [
		_endpoint,
		_protocol,
		timeout_ms,
		_reconnect_max_attempts,
		_reconnect_delay_ms,
		_reconnect_timeout_ms,
		_heartbeat_interval_ms,
		_heartbeat_timeout_ms,
	])
	connecting.emit()

	var result := _dial_conn(_endpoint, _protocol)
	if result.has("error"):
		_fail_connect(result["error"], result["message"])
		return false

	_conn = result["conn"]
	return true

func _fail_connect(error: int, error_message: String) -> void:
	_logger.debug("connect failed error=%d message=%s", [error, error_message])
	_reset_for_connect_failure(error, error_message)
	if _conn != null:
		_conn.close()
		_conn = null
	connect_completed.emit(error, error_message)

func _begin_reconnect() -> bool:
	_logger.debug("reconnect begin attempt=%d/%d endpoint=%s", [_reconnect_attempts, _reconnect_max_attempts, _endpoint])
	if _conn != null:
		_conn.close()
		_conn = null

	var result := _dial_conn(_endpoint, _protocol)
	if result.has("error"):
		_reconnect_error = result["error"]
		_reconnect_error_message = result["message"]
		_logger.debug("reconnect dial failed attempt=%d/%d error=%d message=%s", [
			_reconnect_attempts,
			_reconnect_max_attempts,
			_reconnect_error,
			_reconnect_error_message,
		])
		return false

	_conn = result["conn"]
	_reset_for_reconnect_attempt(RECONNECT_CONNECTING)
	return true

func _fail_reconnect(error: int, error_message: String) -> void:
	_logger.debug("reconnect failed attempt=%d/%d error=%d message=%s", [_reconnect_attempts, _reconnect_max_attempts, error, error_message])
	reconnect_completed.emit(_reconnect_attempts, _reconnect_max_attempts, error, error_message)
	if _reconnect_attempts >= _reconnect_max_attempts:
		close(error, error_message)
		return
	_schedule_reconnect(error, error_message)

func _schedule_reconnect(error: int, error_message: String) -> void:
	_reconnect_error = error
	_reconnect_error_message = error_message if error_message != "" else "connection closed"
	_next_reconnect_ms = _local_unix_time_ms() + _reconnect_delay_ms
	_logger.debug("reconnect scheduled attempt=%d/%d delay_ms=%d error=%d message=%s", [
		_reconnect_attempts + 1,
		_reconnect_max_attempts,
		_reconnect_delay_ms,
		error,
		_reconnect_error_message,
	])
	if _conn != null:
		_conn.close()
		_conn = null
	_reset_for_reconnect_attempt(RECONNECT_WAITING)

func _dial_conn(endpoint: String, protocol: int) -> Dictionary:
	match protocol:
		PROTOCOL_TCP:
			var parsed := _parse_tcp_endpoint(endpoint)
			if parsed.has("error"):
				return {
					"error": ERR_INVALID_PARAMETER,
					"message": parsed["error"],
				}
			var tcp_conn := GolaxyTCPConn.connect_to(parsed["host"], parsed["port"], true, _logger)
			if tcp_conn.get_error() != OK:
				return {
					"error": tcp_conn.get_error(),
					"message": tcp_conn.get_error_message(),
				}
			return {"conn": tcp_conn}
		PROTOCOL_WEBSOCKET:
			var ws_conn := GolaxyWebSocketConn.connect_to(
				endpoint,
				null,
				PackedStringArray(),
				PackedStringArray(),
				16 * 1024 * 1024,
				16 * 1024 * 1024,
				4096,
				true,
				_logger
			)
			if ws_conn.get_error() != OK:
				return {
					"error": ws_conn.get_error(),
					"message": ws_conn.get_error_message(),
				}
			return {"conn": ws_conn}
		_:
			return {
				"error": ERR_INVALID_PARAMETER,
				"message": "unsupported protocol: %d" % protocol,
			}

#endregion

#region Send

func _send_hello(session_id: String = "") -> bool:
	var hello := GTPMessages.MsgHello.new()
	hello.version = GTPMessages.VERSION_V1_0
	hello.session_id = session_id
	hello.random = Crypto.new().generate_random_bytes(32)
	hello.cipher_suite = GTPMessages.CipherSuite.new()
	hello.compression = GTPMessages.COMPRESSION_NONE
	return _send_message(hello)

func _send_auth() -> bool:
	var auth := GTPMessages.MsgAuth.new()
	auth.user_id = _user_id
	auth.token = _token
	auth.extensions = PackedByteArray()
	return _send_message(auth)

func _send_continue() -> bool:
	var msg := GTPMessages.MsgContinue.new()
	msg.send_seq = _send_seq
	msg.recv_seq = _recv_seq
	return _send_message(msg)

func _send_probe_time(timeout_ms: int) -> Future:
	var future := new_future(timeout_ms)
	if future == null:
		return null
	var msg := GTPMessages.MsgSyncTime.new()
	msg.corr_id = future.corr_id
	msg.origin_time = _local_unix_time_ms()
	if not _send_reliable_message(msg, GTPMessages.FLAG_REQ_TIME):
		future.call_deferred("reject", ERR_CONNECTION_ERROR, "send probe time failed")
	return future

func _send_message(msg: GTPMessages.Msg, flags: int = 0) -> bool:
	var packet := GTPCodec.encode_packet(msg, flags, 0, 0)
	if packet.is_empty():
		return false
	if _conn == null or _conn.get_status() != GolaxyConn.CONNECTED:
		return false
	return _conn.send_bytes(packet) == OK

func _send_reliable_message(msg: GTPMessages.Msg, flags: int = 0) -> bool:
	var packet := GTPCodec.encode_packet(msg, flags, _send_seq, _recv_seq)
	if packet.is_empty():
		return false
	_send_window.append(Frame.new(_send_seq, packet))
	_send_seq = _next_seq(_send_seq)
	if not _flush_send_window():
		if _is_reconnecting():
			return true
		if _phase == PHASE_CONNECTED:
			_schedule_reconnect(_conn_error_code(), _conn_error_message())
			return true
		close(_conn_error_code(), _conn_error_message())
		return false
	return true

func _flush_send_window() -> bool:
	if _phase != PHASE_CONNECTED:
		return false
	if _conn == null or _conn.get_status() != GolaxyConn.CONNECTED:
		return false
	var now_ms := _local_unix_time_ms()
	var index := 0
	while index < _send_window.size():
		var frame := _send_window[index]
		if frame.offset >= frame.data.size():
			index += 1
			continue
		var sent := _conn.send_partial_bytes(frame.data.slice(frame.offset))
		if sent < 0:
			return false
		if sent > 0 and frame.sent_ms <= 0:
			frame.sent_ms = now_ms
		frame.offset += sent
		if frame.offset < frame.data.size():
			return true
		index += 1
	return true

#endregion

#region Poll Steps

func _process_connect_timeout() -> void:
	if not _is_connecting():
		return
	if _connect_deadline_ms <= 0:
		return
	if _local_unix_time_ms() < _connect_deadline_ms:
		return
	if _is_reconnecting():
		_fail_reconnect(ERR_TIMEOUT, "reconnect timeout")
	else:
		_fail_connect(ERR_TIMEOUT, "connect timeout")

func _process_phase() -> void:
	if _conn == null:
		return
	match _phase:
		PHASE_CONNECTING:
			match _conn.get_status():
				GolaxyConn.CONNECTED:
					_phase = PHASE_WAIT_HELLO
					var hello_session_id := _session_id if _is_reconnecting() else ""
					if not _send_hello(hello_session_id):
						_handle_connection_lost(_conn_error_code(), _conn_error_message())
						return
				GolaxyConn.CLOSED:
					_handle_connection_lost(_conn_error_code(), _conn_error_message())
					return
		PHASE_WAIT_HELLO, PHASE_WAIT_FINISHED:
			if _conn.get_status() == GolaxyConn.CLOSED:
				_handle_connection_lost(_conn_error_code(), _conn_error_message())
				return
		PHASE_CONNECTED:
			if _conn.get_status() == GolaxyConn.CLOSED:
				_handle_connection_lost(_conn_error_code(), _conn_error_message())
				return

func _process_io() -> void:
	if _conn == null:
		return
	var conn_status := _conn.get_status()
	if conn_status != GolaxyConn.CONNECTED and not _conn.has_pending_bytes():
		return
	if _phase == PHASE_CONNECTED and conn_status == GolaxyConn.CONNECTED and not _flush_send_window():
		_handle_connection_lost(_conn_error_code(), _conn_error_message())
		return
	var bytes := _conn.read_bytes()
	if _conn.get_error() != OK:
		_handle_connection_lost(_conn_error_code(), _conn_error_message())
		return
	if not bytes.is_empty():
		_mark_recv_activity()
		_recv_buff.append_array(bytes)
		_process_recv_buffer()

func _process_heartbeat() -> void:
	if _phase != PHASE_CONNECTED:
		return
	if _conn == null or _conn.get_status() != GolaxyConn.CONNECTED:
		return
	var now_ms := _local_unix_time_ms()
	if _last_recv_ms <= 0:
		_last_recv_ms = now_ms
	if _heartbeat_timeout_ms > 0 and now_ms - _last_recv_ms >= _heartbeat_timeout_ms:
		_logger.debug("heartbeat timeout elapsed_ms=%d timeout_ms=%d send_window=%d", [now_ms - _last_recv_ms, _heartbeat_timeout_ms, _send_window.size()])
		_handle_connection_lost(ERR_TIMEOUT, "heartbeat timeout")
		return
	if _heartbeat_interval_ms <= 0:
		return
	if now_ms - _last_recv_ms < _heartbeat_interval_ms or now_ms - _last_ping_ms < _heartbeat_interval_ms:
		return
	if _send_reliable_message(GTPMessages.MsgHeartbeat.new(), GTPMessages.FLAG_PING):
		_last_ping_ms = now_ms
		_logger.debug("heartbeat ping sent send_seq=%d recv_seq=%d send_window=%d", [_send_seq, _recv_seq, _send_window.size()])

func _process_reconnect() -> void:
	if _reconnect_state != RECONNECT_WAITING:
		return
	if _reconnect_attempts >= _reconnect_max_attempts:
		close(_reconnect_error, _reconnect_error_message)
		return
	if _local_unix_time_ms() < _next_reconnect_ms:
		return
	_reconnect_attempts += 1
	_logger.debug("reconnecting attempt=%d/%d", [_reconnect_attempts, _reconnect_max_attempts])
	reconnecting.emit(_reconnect_attempts, _reconnect_max_attempts)
	if not _begin_reconnect():
		reconnect_completed.emit(_reconnect_attempts, _reconnect_max_attempts, _reconnect_error, _reconnect_error_message)
		if _reconnect_attempts < _reconnect_max_attempts:
			_schedule_reconnect(_reconnect_error, _reconnect_error_message)
			return
		close(_reconnect_error, _reconnect_error_message)
		return

func _process_recv_buffer() -> void:
	while true:
		var length := GTPCodec.peek_packet_length(_recv_buff)
		if length < 0:
			close(ERR_INVALID_DATA, "decode GTP packet failed")
			return
		if length == 0 or _recv_buff.size() < length:
			return
		var frame := _recv_buff.slice(0, length)
		_recv_buff = _recv_buff.slice(length)
		var packet := GTPCodec.decode_packet(frame)
		if packet == null:
			close(ERR_INVALID_DATA, "decode GTP packet failed")
			return
		match _phase:
			PHASE_WAIT_HELLO, PHASE_WAIT_FINISHED:
				_handle_handshake_packet(packet)
			PHASE_CONNECTED:
				if _accept_connected_packet(packet):
					_handle_connected_packet(packet)

func _process_futures() -> void:
	if _futures.is_empty():
		return
	var now_ms := _local_unix_time_ms()
	for corr_id in _futures.keys():
		var future := _futures[corr_id]
		if not future.is_done() and now_ms >= future.deadline_ms:
			_logger.debug("future timeout corr_id=%d", [future.corr_id])
			future.reject(ERR_TIMEOUT, "future timeout")
		if future.is_done():
			_futures.erase(corr_id)

#endregion

#region Packet Handlers

func _handle_handshake_packet(packet: GTPMessages.MsgPacket) -> void:
	var head := packet.head
	match head.msg_id:
		GTPMessages.MSG_ID_RST:
			var msg := packet.body as GTPMessages.MsgRst
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgRst")
			close(ERR_CONNECTION_ERROR, "server reset (%d): %s" % [msg.code, msg.message])
		GTPMessages.MSG_ID_HELLO:
			if _phase != PHASE_WAIT_HELLO:
				close(ERR_INVALID_DATA, "unexpected GTP hello packet")
				return
			if (head.flags & GTPMessages.FLAG_HELLO_DONE) == 0:
				close(ERR_INVALID_DATA, "server GTP hello packet is missing hello-done flag")
				return
			var msg := packet.body as GTPMessages.MsgHello
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgHello")
			_logger.debug("handshake hello received reconnect=%s flags=%d session=%s send_seq=%d recv_seq=%d", [
				_is_reconnecting(),
				head.flags,
				msg.session_id,
				_send_seq,
				_recv_seq,
			])
			var error := GTPCodec.supports_hello(msg)
			if error != "":
				close(ERR_INVALID_DATA, error)
				return
			if _is_reconnecting():
				if (head.flags & GTPMessages.FLAG_CONTINUE) == 0:
					close(ERR_INVALID_DATA, "server reconnect GTP hello packet is missing continue flag")
					return
				if msg.session_id != _session_id:
					close(ERR_INVALID_DATA, "reconnect session mismatch")
					return
			else:
				_session_id = msg.session_id
				if (head.flags & GTPMessages.FLAG_CONTINUE) != 0:
					close(ERR_UNAVAILABLE, "unexpected session continue")
					return
			_expect_auth = (head.flags & GTPMessages.FLAG_AUTH) != 0
			_phase = PHASE_WAIT_FINISHED
			if _expect_auth:
				_logger.debug("handshake auth sending")
				if not _send_auth():
					if _is_reconnecting():
						_fail_reconnect(_conn_error_code(), _conn_error_message())
						return
					close(_conn_error_code(), _conn_error_message())
					return
			if _is_reconnecting():
				_logger.debug("handshake continue sending session=%s send_seq=%d recv_seq=%d send_window=%d", [
					_session_id,
					_send_seq,
					_recv_seq,
					_send_window.size(),
				])
				if not _send_continue():
					_fail_reconnect(_conn_error_code(), _conn_error_message())
		GTPMessages.MSG_ID_FINISHED:
			if _phase != PHASE_WAIT_FINISHED:
				close(ERR_INVALID_DATA, "unexpected GTP finished packet")
				return
			if _expect_auth and (head.flags & GTPMessages.FLAG_AUTH_OK) == 0:
				close(ERR_UNAUTHORIZED, "server GTP finished packet is missing auth-ok flag")
				return
			_expect_auth = false
			var msg := packet.body as GTPMessages.MsgFinished
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgFinished")
			_logger.debug("handshake finished received reconnect=%s flags=%d remote_send_seq=%d remote_recv_seq=%d local_send_seq=%d local_recv_seq=%d", [
				_is_reconnecting(),
				head.flags,
				msg.send_seq,
				msg.recv_seq,
				_send_seq,
				_recv_seq,
			])
			if _is_reconnecting():
				if (head.flags & GTPMessages.FLAG_CONTINUE_OK) == 0:
					close(ERR_INVALID_DATA, "server reconnect GTP finished packet is missing continue-ok flag")
					return
				if not _synchronize_send_window(msg.recv_seq):
					close(ERR_CONNECTION_ERROR, "reconnect resend cache is out of date")
					return
				var reconnect_attempts := _reconnect_attempts
				_reset_reconnect_state()
				_reset_heartbeat_state()
				_phase = PHASE_CONNECTED
				_logger.debug("reconnect ok attempt=%d/%d session=%s send_seq=%d recv_seq=%d send_window=%d", [
					reconnect_attempts,
					_reconnect_max_attempts,
					_session_id,
					_send_seq,
					_recv_seq,
					_send_window.size(),
				])
				reconnect_completed.emit(reconnect_attempts, _reconnect_max_attempts, OK, "")
				reconnected.emit(reconnect_attempts, _reconnect_max_attempts)
			else:
				_send_seq = msg.recv_seq
				_recv_seq = msg.send_seq
				_connect_deadline_ms = 0
				_reset_heartbeat_state()
				_phase = PHASE_CONNECTED
				_logger.debug("connect ok session=%s send_seq=%d recv_seq=%d", [_session_id, _send_seq, _recv_seq])
				connect_completed.emit(OK, "")
				connected.emit()
		_:
			_logger.warning("unexpected GTP packet during handshake: %d", [head.msg_id])

func _handle_connected_packet(packet: GTPMessages.MsgPacket) -> void:
	var head := packet.head
	match head.msg_id:
		GTPMessages.MSG_ID_RST:
			var msg := packet.body as GTPMessages.MsgRst
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgRst")
			close(ERR_CONNECTION_ERROR, "server reset (%d): %s" % [msg.code, msg.message])
		GTPMessages.MSG_ID_SYNC_TIME:
			var msg := packet.body as GTPMessages.MsgSyncTime
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgSyncTime")
			_handle_sync_time(head, msg)
		GTPMessages.MSG_ID_HEARTBEAT:
			_handle_heartbeat(head)
		GTPMessages.MSG_ID_PAYLOAD:
			var msg := packet.body as GTPMessages.MsgPayload
			assert(msg != null, "incorrect message type; expected GTPMessages.MsgPayload")
			data_received.emit(msg.data)
		_:
			_logger.warning("unexpected GTP packet while connected: %d", [head.msg_id])

func _handle_sync_time(head: GTPMessages.MsgHead, msg: GTPMessages.MsgSyncTime) -> void:
	if (head.flags & GTPMessages.FLAG_REQ_TIME) != 0:
		var receive_ms := _local_unix_time_ms()
		var resp := GTPMessages.MsgSyncTime.new()
		resp.corr_id = msg.corr_id
		resp.origin_time = msg.origin_time
		resp.receive_time = receive_ms
		resp.transmit_time = _local_unix_time_ms()
		resp.zone_offset = _local_zone_offset()
		if not _send_reliable_message(resp, GTPMessages.FLAG_RESP_TIME):
			_logger.error("failed to send GTP sync-time response")
		return

	if (head.flags & GTPMessages.FLAG_RESP_TIME) == 0:
		_logger.warning("unexpected GTP sync-time packet without req-time or resp-time flag")
		return

	var future := get_future(msg.corr_id)
	if future == null:
		_logger.warning("unexpected GTP sync-time response corr_id: %d", [msg.corr_id])
		return

	var time_sample := TimeSample.new()
	time_sample.origin_ms = msg.origin_time
	time_sample.receive_ms = msg.receive_time
	time_sample.transmit_ms = msg.transmit_time
	time_sample.destination_ms = _local_unix_time_ms()
	time_sample.remote_zone_offset = msg.zone_offset
	future.resolve(time_sample)

func _handle_heartbeat(head: GTPMessages.MsgHead) -> void:
	if (head.flags & GTPMessages.FLAG_PING) != 0:
		_logger.debug("heartbeat ping received seq=%d ack=%d send_seq=%d recv_seq=%d send_window=%d", [
			head.seq,
			head.ack,
			_send_seq,
			_recv_seq,
			_send_window.size(),
		])
		if not _send_reliable_message(GTPMessages.MsgHeartbeat.new(), GTPMessages.FLAG_PONG):
			_logger.error("failed to send GTP heartbeat pong")
	elif (head.flags & GTPMessages.FLAG_PONG) != 0:
		_logger.debug("heartbeat pong received seq=%d ack=%d send_seq=%d recv_seq=%d send_window=%d", [
			head.seq,
			head.ack,
			_send_seq,
			_recv_seq,
			_send_window.size(),
		])
	else:
		_logger.warning("unexpected GTP heartbeat packet without ping or pong flag")

func _accept_connected_packet(packet: GTPMessages.MsgPacket) -> bool:
	var seq := int(packet.head.seq)
	if seq != _recv_seq:
		if _seq_before(seq, _recv_seq):
			_logger.warning("discard stale GTP packet sequence: %d, expected: %d", [seq, _recv_seq])
			return false
		close(ERR_INVALID_DATA, "unexpected GTP sequence: %d, expected: %d" % [seq, _recv_seq])
		return false
	var ack := int(packet.head.ack)
	if _seq_before(_send_seq, ack):
		close(ERR_INVALID_DATA, "unexpected GTP ack: %d, next send sequence: %d" % [ack, _send_seq])
		return false
	_ack_send_window(ack)
	_recv_seq = _next_seq(_recv_seq)
	return true

#endregion

#region Utilities

func _conn_error_code() -> int:
	if _conn != null and _conn.get_error() != OK:
		return _conn.get_error()
	return ERR_CONNECTION_ERROR

func _conn_error_message() -> String:
	if _conn != null and _conn.get_error_message() != "":
		return _conn.get_error_message()
	return "connection closed"

func _handle_connection_lost(error: int, error_message: String) -> void:
	_logger.debug("connection lost phase=%d reconnect_state=%d error=%d message=%s conn_status=%d conn_error=%d conn_message=%s", [
		_phase,
		_reconnect_state,
		error,
		error_message,
		_conn.get_status() if _conn != null else -1,
		_conn.get_error() if _conn != null else OK,
		_conn.get_error_message() if _conn != null else "",
	])
	if _is_reconnecting():
		_fail_reconnect(error, error_message)
	elif _phase == PHASE_CONNECTED:
		_schedule_reconnect(error, error_message)
	else:
		close(error, error_message)

func _reset_for_initial_connect(timeout_ms: int) -> void:
	_phase = PHASE_CONNECTING
	_connect_deadline_ms = _local_unix_time_ms() + maxi(timeout_ms, 0)
	_close_error = OK
	_close_error_message = ""
	_reset_session_state()
	_reset_reconnect_state()

func _reset_for_reconnect_attempt(reconnect_state: int) -> void:
	_phase = PHASE_CONNECTING
	_reconnect_state = reconnect_state
	_connect_deadline_ms = _local_unix_time_ms() + maxi(_reconnect_timeout_ms, 0) if reconnect_state == RECONNECT_CONNECTING else 0
	_reset_handshake_state()

func _reset_for_connect_failure(error: int, error_message: String) -> void:
	_phase = PHASE_CLOSED
	_connect_deadline_ms = 0
	_close_error = error
	_close_error_message = error_message

func _reset_for_close(error: int, error_message: String) -> void:
	_phase = PHASE_CLOSED
	_connect_deadline_ms = 0
	_close_error = error
	_close_error_message = error_message
	_reset_session_state()
	_reset_reconnect_state()

func _reset_session_state() -> void:
	_send_window.clear()
	_send_seq = 0
	_recv_seq = 0
	_reset_heartbeat_state()
	_reset_handshake_state()

func _reset_handshake_state() -> void:
	_recv_buff = PackedByteArray()
	_expect_auth = false

func _reset_reconnect_state() -> void:
	_reconnect_state = RECONNECT_IDLE
	_reconnect_attempts = 0
	_next_reconnect_ms = 0
	_reconnect_error = OK
	_reconnect_error_message = ""

func _reset_heartbeat_state() -> void:
	_last_recv_ms = _local_unix_time_ms()
	_last_ping_ms = 0

func _mark_recv_activity() -> void:
	_last_recv_ms = _local_unix_time_ms()
	_last_ping_ms = 0

func _ack_send_window(ack: int) -> void:
	while not _send_window.is_empty():
		if not _seq_before(_send_window[0].seq, ack):
			return
		_send_window.pop_front()

func _synchronize_send_window(remote_recv_seq: int) -> bool:
	var replay_seq := remote_recv_seq
	if remote_recv_seq == _send_seq and not _send_window.is_empty():
		replay_seq = _prev_seq(remote_recv_seq)
	_logger.debug("synchronize send window remote_recv_seq=%d replay_seq=%d send_seq=%d cached=%d", [remote_recv_seq, replay_seq, _send_seq, _send_window.size()])
	_ack_send_window(replay_seq)
	if _send_window.is_empty():
		return remote_recv_seq == _send_seq
	if _send_window[0].seq != replay_seq:
		return false
	for frame in _send_window:
		frame.offset = 0
		frame.sent_ms = 0
	return true

func _is_connecting() -> bool:
	return _phase == PHASE_CONNECTING or _phase == PHASE_WAIT_HELLO or _phase == PHASE_WAIT_FINISHED

func _is_reconnecting() -> bool:
	return _reconnect_state != RECONNECT_IDLE

#endregion

#region Static Helpers

static func _parse_tcp_endpoint(endpoint: String) -> Dictionary:
	var value := endpoint
	if value.begins_with("tcp://"):
		value = value.substr(6)
	var slash_idx := value.find("/")
	if slash_idx >= 0:
		value = value.substr(0, slash_idx)
	if value == "":
		return {"error": "tcp endpoint is empty"}

	var host := ""
	var port_text := ""

	if value.begins_with("["):
		var end_idx := value.find("]")
		if end_idx < 0 or end_idx + 1 >= value.length() or value[end_idx + 1] != ":":
			return {"error": "invalid tcp endpoint: %s" % endpoint}
		host = value.substr(1, end_idx - 1)
		port_text = value.substr(end_idx + 2)
	else:
		var colon_idx := value.rfind(":")
		if colon_idx <= 0 or colon_idx >= value.length() - 1:
			return {"error": "invalid tcp endpoint: %s" % endpoint}
		host = value.substr(0, colon_idx)
		port_text = value.substr(colon_idx + 1)

	if not port_text.is_valid_int():
		return {"error": "invalid tcp endpoint port: %s" % endpoint}

	var port := int(port_text)
	if port < 0 or port > 65535:
		return {"error": "tcp endpoint port out of range: %s" % endpoint}

	return {
		"host": host,
		"port": port,
	}

static func _next_seq(seq: int) -> int:
	return (seq + 1) & 0xffffffff

static func _prev_seq(seq: int) -> int:
	return (seq - 1) & 0xffffffff

static func _seq_before(a: int, b: int) -> bool:
	var diff := (a - b) & 0xffffffff
	return diff != 0 and diff >= 0x80000000

static func _local_unix_time_ms() -> int:
	return int(Time.get_unix_time_from_system() * 1000.0)

static func _local_zone_offset() -> int:
	var zone := Time.get_time_zone_from_system()
	return int(zone.get("bias", 0)) * 60

#endregion
