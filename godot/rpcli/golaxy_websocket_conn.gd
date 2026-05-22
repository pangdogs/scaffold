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
class_name GolaxyWebSocketConn
extends GolaxyConn

var _peer: WebSocketPeer = WebSocketPeer.new()
var _error: int = OK
var _error_message: String = ""
var _no_delay: bool = true
var _no_delay_applied: bool = false
var _last_ready_state: int = -1
var _logger := GolaxyLogger.new("GolaxyWebSocketConn", get_instance_id())

var logger: GolaxyLogger:
	get:
		return _logger

static func connect_to(
	url: String,
	tls_options: TLSOptions = null,
	handshake_headers: PackedStringArray = PackedStringArray(),
	supported_protocols: PackedStringArray = PackedStringArray(),
	inbound_buffer_size: int = 16 * 1024 * 1024,
	outbound_buffer_size: int = 16 * 1024 * 1024,
	max_queued_packets: int = 4096,
	no_delay: bool = true,
	logger: GolaxyLogger = null
) -> GolaxyWebSocketConn:
	var conn := GolaxyWebSocketConn.new()
	if logger != null:
		conn._logger = logger.named("GolaxyWebSocketConn", conn.get_instance_id())
	conn._peer.handshake_headers = handshake_headers
	conn._peer.supported_protocols = supported_protocols
	conn._peer.inbound_buffer_size = inbound_buffer_size
	conn._peer.outbound_buffer_size = outbound_buffer_size
	conn._peer.max_queued_packets = max_queued_packets
	conn._connect(url, tls_options, no_delay)
	return conn

func poll() -> void:
	_peer.poll()
	_trace_ready_state()
	_apply_no_delay()

func get_status() -> int:
	match _peer.get_ready_state():
		WebSocketPeer.STATE_CONNECTING:
			return GolaxyConn.CONNECTING
		WebSocketPeer.STATE_OPEN:
			return GolaxyConn.CONNECTED
		_:
			return GolaxyConn.CLOSED

func send_bytes(data: PackedByteArray) -> int:
	_clear_error()
	if data.is_empty():
		return OK
	var state := _peer.get_ready_state()
	if state != WebSocketPeer.STATE_OPEN:
		_fail(ERR_CONNECTION_ERROR, "websocket connection state: %d" % state)
		return ERR_CONNECTION_ERROR
	var error := _peer.send(data, WebSocketPeer.WRITE_MODE_BINARY)
	if error != OK:
		_fail(error, "websocket connection send packet failed")
	return error

func send_partial_bytes(data: PackedByteArray) -> int:
	var error := send_bytes(data)
	if error != OK:
		return -1
	return data.size()

func read_bytes() -> PackedByteArray:
	_clear_error()
	var bytes := PackedByteArray()
	while _peer.get_available_packet_count() > 0:
		bytes.append_array(_peer.get_packet())
		if _peer.get_packet_error() != OK:
			_fail(_peer.get_packet_error(), "websocket connection read packet failed")
			return PackedByteArray()
	return bytes

func has_pending_bytes() -> bool:
	return _peer.get_available_packet_count() > 0

func close() -> void:
	_logger.debug("close")
	_peer.close()

func get_error() -> int:
	return _error

func get_error_message() -> String:
	return _error_message

func _connect(url: String, tls_client_options: TLSOptions, no_delay: bool) -> int:
	_no_delay = no_delay
	_no_delay_applied = false
	_last_ready_state = -1
	_logger.debug("connect url=%s no_delay=%s", [url, no_delay])
	var error := _peer.connect_to_url(url, tls_client_options)
	if error != OK:
		_fail(error, "websocket connect failed: %s" % url)
		return error
	return OK

func _apply_no_delay() -> void:
	if _no_delay_applied:
		return
	if _peer.get_ready_state() != WebSocketPeer.STATE_OPEN:
		return
	if not OS.has_feature("web"):
		_peer.set_no_delay(_no_delay)
	_no_delay_applied = true

func _clear_error() -> void:
	_error = OK
	_error_message = ""

func _fail(error: int, message: String) -> void:
	_error = error
	_error_message = message
	_logger.debug("fail error=%d message=%s", [error, message])

func _trace_ready_state() -> void:
	var state := _peer.get_ready_state()
	if state == _last_ready_state:
		return
	_last_ready_state = state
	var close_code := _peer.get_close_code()
	var close_reason := _peer.get_close_reason()
	_logger.debug("ready_state=%d close_code=%d close_reason=%s", [state, close_code, close_reason])
