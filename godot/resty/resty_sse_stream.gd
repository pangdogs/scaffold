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
class_name RestySSEStream
extends Node

signal opened()
signal event_received(event: RestySSEEvent)
signal closed(error_message: String)

enum State {
	IDLE,
	CONNECTING,
	REQUESTING,
	OPEN,
	CLOSED,
}

var _client: RestyClient = null
var _http := HTTPClient.new()
var _url := ""
var _method: int = HTTPClient.METHOD_GET
var _base_url := ""
var _http_options: RestyHttpOptions = null
var _headers := {}
var _query_params := {}
var _path_params := {}
var _body: Variant = null
var _body_format := RestyRequest.BodyFormat.AUTO
var _body_content_type := ""
var _state := State.IDLE
var _started_at := 0
var _request_header_lines: PackedStringArray = PackedStringArray()
var _request_body: Variant = ""
var _request_path := "/"
var _status_code := 0
var _response_headers := {}
var _response_header_lines: PackedStringArray = PackedStringArray()
var _pending := PackedByteArray()
var _last_event_id := ""
var _error_message := ""

var status_code: int:
	get:
		return _status_code

var response_headers: Dictionary:
	get:
		return _response_headers

var response_header_lines: PackedStringArray:
	get:
		return _response_header_lines

var last_event_id: String:
	get:
		return _last_event_id

var error_message: String:
	get:
		return _error_message

func _init(client: RestyClient, base_url: String, http_options: RestyHttpOptions, headers: Dictionary, query_params: Dictionary) -> void:
	_client = client
	_base_url = base_url
	_http_options = http_options.duplicate()
	_headers = headers.duplicate(true)
	_query_params = query_params.duplicate(true)

func _ready() -> void:
	set_process(false)

func _process(_delta: float) -> void:
	if _state == State.IDLE or _state == State.CLOSED:
		return

	if _http_options.timeout > 0.0 and _state != State.OPEN:
		var elapsed := float(Time.get_ticks_msec() - _started_at) / 1000.0
		if elapsed >= _http_options.timeout:
			_fail("sse timeout")
			return

	var err := _http.poll()
	if err != OK:
		if _state == State.OPEN:
			close()
			return
		_fail("sse poll failed: %s" % error_string(err))
		return

	match _http.get_status():
		HTTPClient.STATUS_RESOLVING, HTTPClient.STATUS_CONNECTING, HTTPClient.STATUS_REQUESTING:
			return
		HTTPClient.STATUS_CONNECTED:
			if _state == State.CONNECTING:
				_send_request()
			return
		HTTPClient.STATUS_BODY:
			_handle_body()
		HTTPClient.STATUS_DISCONNECTED:
			close()
		_:
			_fail("sse connection status: %d" % _http.get_status())

func set_header(name: String, value: Variant) -> RestySSEStream:
	if _state != State.IDLE:
		return self
	RestyClient._set_header(_headers, name, str(value))
	return self

func set_headers(values: Dictionary) -> RestySSEStream:
	if _state != State.IDLE:
		return self
	for key in values:
		set_header(str(key), values[key])
	return self

func set_timeout(seconds: float) -> RestySSEStream:
	if _state != State.IDLE:
		return self
	_http_options.timeout = seconds
	return self

func set_bearer_auth(token: String) -> RestySSEStream:
	return set_header("Authorization", "Bearer %s" % token)

func set_basic_auth(username: String, password: String) -> RestySSEStream:
	var token := Marshalls.utf8_to_base64("%s:%s" % [username, password])
	return set_header("Authorization", "Basic %s" % token)

func set_query_param(name: String, value: Variant) -> RestySSEStream:
	if _state != State.IDLE:
		return self
	_query_params[name] = value
	return self

func set_query_params(values: Dictionary) -> RestySSEStream:
	if _state != State.IDLE:
		return self
	for key in values:
		set_query_param(str(key), values[key])
	return self

func set_path_param(name: String, value: Variant) -> RestySSEStream:
	if _state != State.IDLE:
		return self
	_path_params[name] = value
	return self

func set_path_params(values: Dictionary) -> RestySSEStream:
	if _state != State.IDLE:
		return self
	for key in values:
		set_path_param(str(key), values[key])
	return self

func set_content_type(value: String) -> RestySSEStream:
	return set_header("Content-Type", value)

func set_body(value: Variant) -> RestySSEStream:
	if _state != State.IDLE:
		return self
	_body = value
	_body_format = RestyRequest.BodyFormat.AUTO
	_body_content_type = ""
	return self

func set_raw_body(value: PackedByteArray, content_type: String = "application/octet-stream") -> RestySSEStream:
	if _state != State.IDLE:
		return self
	_body = value
	_body_format = RestyRequest.BodyFormat.RAW
	_body_content_type = content_type
	return self

func set_json(value: Variant, content_type: String = "application/json") -> RestySSEStream:
	if _state != State.IDLE:
		return self
	_body = value
	_body_format = RestyRequest.BodyFormat.JSON
	_body_content_type = content_type
	return self

func set_form(values: Dictionary, content_type: String = "application/x-www-form-urlencoded") -> RestySSEStream:
	if _state != State.IDLE:
		return self
	_body = values
	_body_format = RestyRequest.BodyFormat.FORM
	_body_content_type = content_type
	return self

func start(method: int, url: String) -> bool:
	if _state != State.IDLE:
		return false
	_method = method
	_url = url
	_state = State.CONNECTING
	_started_at = Time.get_ticks_msec()

	var parsed := _parse_url(RestyClient._build_url(self))
	if parsed.is_empty():
		_fail("invalid url: %s" % _url)
		return false

	_request_path = parsed["path"]
	var headers := _headers.duplicate()
	RestyClient._set_header_if_missing(headers, "Accept", "text/event-stream")
	RestyClient._set_header_if_missing(headers, "Cache-Control", "no-cache")
	_request_body = RestyClient._build_body(self, headers)
	_request_header_lines = RestyClient._build_headers(headers)

	var tls_options = TLSOptions.client() if bool(parsed["tls"]) else null
	var err := _http.connect_to_host(parsed["host"], int(parsed["port"]), tls_options)
	if err != OK:
		_fail("connect failed: %s" % error_string(err))
		return false

	set_process(true)
	return true

func close() -> void:
	if _state == State.CLOSED:
		return
	_state = State.CLOSED
	set_process(false)
	_http.close()
	call_deferred("_emit_closed")

func get_response_header(name: String) -> String:
	return str(_response_headers.get(RestyClient._canonical_header_name(name), ""))

func _send_request() -> void:
	var err := OK
	if _request_body is PackedByteArray:
		err = _http.request_raw(_method, _request_path, _request_header_lines, _request_body)
	else:
		err = _http.request(_method, _request_path, _request_header_lines, str(_request_body))
	if err != OK:
		_fail("sse request failed: %s" % error_string(err))
		return
	_state = State.REQUESTING

func _handle_body() -> void:
	if _state != State.OPEN:
		_status_code = _http.get_response_code()
		_response_header_lines = _http.get_response_headers()
		_response_headers = RestyClient._parse_headers(_response_header_lines)
		if _status_code < 200 or _status_code >= 300:
			_fail("http status: %d" % _status_code)
			return
		_state = State.OPEN
		opened.emit()

	var chunk := _http.read_response_body_chunk()
	if chunk.is_empty():
		return
	_pending.append_array(chunk)
	_parse_pending_events()

func _parse_pending_events() -> void:
	while true:
		var event_bytes: Variant = _pop_pending_event()
		if event_bytes == null:
			return
		_parse_event(event_bytes.get_string_from_utf8())

func _pop_pending_event() -> Variant:
	for index in range(_pending.size()):
		if _pending[index] == 10:
			if index + 1 < _pending.size() and _pending[index + 1] == 10:
				var event_bytes := _pending.slice(0, index)
				_pending = _pending.slice(index + 2)
				return event_bytes
			if index + 2 < _pending.size() and _pending[index + 1] == 13 and _pending[index + 2] == 10:
				var event_bytes := _pending.slice(0, index)
				_pending = _pending.slice(index + 3)
				return event_bytes
		elif _pending[index] == 13:
			if index + 1 < _pending.size() and _pending[index + 1] == 13:
				var event_bytes := _pending.slice(0, index)
				_pending = _pending.slice(index + 2)
				return event_bytes
			if index + 3 < _pending.size() and _pending[index + 1] == 10 and _pending[index + 2] == 13 and _pending[index + 3] == 10:
				var event_bytes := _pending.slice(0, index)
				_pending = _pending.slice(index + 4)
				return event_bytes
	return null

func _parse_event(text: String) -> void:
	var event_name := ""
	var event_data := PackedStringArray()
	var event_id := ""
	var event_retry := -1

	for raw_line in text.replace("\r\n", "\n").replace("\r", "\n").split("\n"):
		var line := String(raw_line)
		if line.is_empty() or line.begins_with(":"):
			continue

		var separator := line.find(":")
		var field := line
		var value := ""
		if separator >= 0:
			field = line.substr(0, separator)
			value = line.substr(separator + 1)
			if value.begins_with(" "):
				value = value.substr(1)

		match field:
			"event":
				event_name = value
			"data":
				event_data.append(value)
			"id":
				event_id = value
			"retry":
				if value.is_valid_int():
					event_retry = int(value)

	if event_data.is_empty():
		return

	var event := RestySSEEvent.new()
	event.name = "message" if event_name.is_empty() else event_name
	event.data = "\n".join(event_data)
	event.id = event_id
	event.retry = event_retry
	_emit_event(event)

func _emit_event(event: RestySSEEvent) -> void:
	if not event.id.is_empty():
		_last_event_id = event.id
	event_received.emit(event)

func _parse_url(url: String) -> Dictionary:
	var scheme_end := url.find("://")
	if scheme_end < 0:
		return {}

	var scheme := url.substr(0, scheme_end).to_lower()
	if scheme != "http" and scheme != "https":
		return {}

	var rest := url.substr(scheme_end + 3)
	var path_start := rest.find("/")
	var query_start := rest.find("?")
	if path_start < 0 or (query_start >= 0 and query_start < path_start):
		path_start = query_start
	var authority := rest
	var path := "/"
	if path_start >= 0:
		authority = rest.substr(0, path_start)
		path = rest.substr(path_start)
		if path.begins_with("?"):
			path = "/" + path

	if authority.is_empty():
		return {}

	var host := authority
	var port := 443 if scheme == "https" else 80
	if authority.begins_with("["):
		var close_bracket := authority.find("]")
		if close_bracket < 0:
			return {}
		host = authority.substr(1, close_bracket - 1)
		if authority.length() > close_bracket + 1 and authority.substr(close_bracket + 1, 1) == ":":
			port = int(authority.substr(close_bracket + 2))
	else:
		var port_separator := authority.rfind(":")
		if port_separator >= 0:
			host = authority.substr(0, port_separator)
			port = int(authority.substr(port_separator + 1))

	if host.is_empty():
		return {}

	return {
		"host": host,
		"port": port,
		"path": path,
		"tls": scheme == "https",
	}

func _fail(error_message: String) -> void:
	if _state == State.CLOSED:
		return
	_error_message = error_message
	close()

func _emit_closed() -> void:
	event_received.emit(null)
	closed.emit(_error_message)
	queue_free()
