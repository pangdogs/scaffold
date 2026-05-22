class_name GolaxyTCPConn
extends GolaxyConn

var _peer: StreamPeerTCP = StreamPeerTCP.new()
var _error: int = OK
var _error_message: String = ""
var _no_delay: bool = true
var _no_delay_applied: bool = false
var _last_status: int = -1
var _logger := GolaxyLogger.new("GolaxyTCPConn", get_instance_id())

var logger: GolaxyLogger:
	get:
		return _logger

static func connect_to(host: String, port: int, no_delay: bool = true, logger: GolaxyLogger = null) -> GolaxyTCPConn:
	var conn := GolaxyTCPConn.new()
	if logger != null:
		conn._logger = logger.named("GolaxyTCPConn", conn.get_instance_id())
	conn._connect(host, port, no_delay)
	return conn

func poll() -> void:
	_peer.poll()
	_trace_status()
	_apply_no_delay()

func get_status() -> int:
	match _peer.get_status():
		StreamPeerTCP.STATUS_CONNECTING:
			return GolaxyConn.CONNECTING
		StreamPeerTCP.STATUS_CONNECTED:
			return GolaxyConn.CONNECTED
		_:
			return GolaxyConn.CLOSED

func send_bytes(data: PackedByteArray) -> int:
	_clear_error()
	if data.is_empty():
		return OK
	if _peer.get_status() != StreamPeerTCP.STATUS_CONNECTED:
		_fail(ERR_CONNECTION_ERROR, "tcp connection state: %d" % _peer.get_status())
		return ERR_CONNECTION_ERROR
	var error := _peer.put_data(data)
	if error != OK:
		_fail(error, "tcp connection send bytes failed")
	return error

func send_partial_bytes(data: PackedByteArray) -> int:
	_clear_error()
	if data.is_empty():
		return 0
	if _peer.get_status() != StreamPeerTCP.STATUS_CONNECTED:
		_fail(ERR_CONNECTION_ERROR, "tcp connection state: %d" % _peer.get_status())
		return -1
	var result: Array = _peer.put_partial_data(data)
	var error := int(result[0])
	if error != OK:
		_fail(error, "tcp connection send partial bytes failed")
		return -1
	return int(result[1])

func read_bytes() -> PackedByteArray:
	_clear_error()
	if _peer.get_status() != StreamPeerTCP.STATUS_CONNECTED:
		_fail(ERR_CONNECTION_ERROR, "tcp connection state: %d" % _peer.get_status())
		return PackedByteArray()
	var available := _peer.get_available_bytes()
	if available <= 0:
		return PackedByteArray()
	var result: Array = _peer.get_partial_data(available)
	var error := int(result[0])
	if error != OK:
		_fail(error, "tcp connection read bytes failed")
		return PackedByteArray()
	return result[1]

func has_pending_bytes() -> bool:
	return _peer.get_available_bytes() > 0

func close() -> void:
	_logger.debug("close")
	_peer.disconnect_from_host()

func get_error() -> int:
	return _error

func get_error_message() -> String:
	return _error_message

func _connect(host: String, port: int, no_delay: bool) -> int:
	_no_delay = no_delay
	_no_delay_applied = false
	_last_status = -1
	_logger.debug("connect host=%s port=%d no_delay=%s", [host, port, no_delay])
	var error := _peer.connect_to_host(host, port)
	if error != OK:
		_fail(error, "tcp connect failed: %s:%d" % [host, port])
		return error
	return OK

func _apply_no_delay() -> void:
	if _no_delay_applied:
		return
	if _peer.get_status() != StreamPeerTCP.STATUS_CONNECTED:
		return
	_peer.set_no_delay(_no_delay)
	_no_delay_applied = true

func _clear_error() -> void:
	_error = OK
	_error_message = ""

func _fail(error: int, message: String) -> void:
	_error = error
	_error_message = message
	_logger.debug("fail error=%d message=%s", [error, message])

func _trace_status() -> void:
	var status := _peer.get_status()
	if status == _last_status:
		return
	_last_status = status
	_logger.debug("status=%d", [status])
