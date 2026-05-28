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
class_name RestyClient
extends Node

var _base_url: String = ""
var _http_options := RestyHttpOptions.new()
var _default_headers := {}
var _default_query_params := {}

func r() -> RestyRequest:
	return RestyRequest.new(self, _base_url, _http_options, _default_headers, _default_query_params)

func sse() -> RestySSEStream:
	var stream := RestySSEStream.new(self, _base_url, _http_options, _default_headers, _default_query_params)
	add_child(stream)
	return stream

func set_base_url(value: String) -> RestyClient:
	_base_url = value.strip_edges()
	return self

func set_header(name: String, value: Variant) -> RestyClient:
	_set_header(_default_headers, name, str(value))
	return self

func set_headers(values: Dictionary) -> RestyClient:
	for key in values:
		set_header(str(key), values[key])
	return self

func set_timeout(seconds: float) -> RestyClient:
	_http_options.timeout = seconds
	return self

func set_parse_json(enabled: bool) -> RestyClient:
	_http_options.parse_json = enabled
	return self

func set_accept_gzip(enabled: bool) -> RestyClient:
	_http_options.accept_gzip = enabled
	return self

func set_body_size_limit(bytes: int) -> RestyClient:
	_http_options.body_size_limit = bytes
	return self

func set_download_chunk_size(bytes: int) -> RestyClient:
	_http_options.download_chunk_size = bytes
	return self

func set_max_redirects(count: int) -> RestyClient:
	_http_options.max_redirects = count
	return self

func set_use_threads(enabled: bool) -> RestyClient:
	_http_options.use_threads = enabled
	return self

func set_bearer_auth(token: String) -> RestyClient:
	return set_header("Authorization", "Bearer %s" % token)

func set_basic_auth(username: String, password: String) -> RestyClient:
	var token := Marshalls.utf8_to_base64("%s:%s" % [username, password])
	return set_header("Authorization", "Basic %s" % token)

func set_query_param(name: String, value: Variant) -> RestyClient:
	_default_query_params[name] = value
	return self

func set_query_params(values: Dictionary) -> RestyClient:
	for key in values:
		set_query_param(str(key), values[key])
	return self

func get_async(url: String = "") -> RestyResponse:
	return await r().get_async(url)

func get_start(url: String = "") -> RestyRequestHandle:
	return r().get_start(url)

func post_async(url: String = "", body: Variant = null) -> RestyResponse:
	var request := r()
	if body != null:
		request.set_body(body)
	return await request.post_async(url)

func post_start(url: String = "", body: Variant = null) -> RestyRequestHandle:
	var request := r()
	if body != null:
		request.set_body(body)
	return request.post_start(url)

func put_async(url: String = "", body: Variant = null) -> RestyResponse:
	var request := r()
	if body != null:
		request.set_body(body)
	return await request.put_async(url)

func put_start(url: String = "", body: Variant = null) -> RestyRequestHandle:
	var request := r()
	if body != null:
		request.set_body(body)
	return request.put_start(url)

func patch_async(url: String = "", body: Variant = null) -> RestyResponse:
	var request := r()
	if body != null:
		request.set_body(body)
	return await request.patch_async(url)

func patch_start(url: String = "", body: Variant = null) -> RestyRequestHandle:
	var request := r()
	if body != null:
		request.set_body(body)
	return request.patch_start(url)

func delete_async(url: String = "") -> RestyResponse:
	return await r().delete_async(url)

func delete_start(url: String = "") -> RestyRequestHandle:
	return r().delete_start(url)

func head_async(url: String = "") -> RestyResponse:
	return await r().head_async(url)

func head_start(url: String = "") -> RestyRequestHandle:
	return r().head_start(url)

func _request_async(request: RestyRequest) -> RestyResponse:
	var handle := _request_start(request)
	return await handle.completed

func _request_start(request: RestyRequest) -> RestyRequestHandle:
	var handle := RestyRequestHandle.new()
	var response := RestyResponse.new()
	var http := HTTPRequest.new()
	add_child(http)

	var headers := request._headers.duplicate()
	var request_body: Variant = _build_body(request, headers)
	var url := _build_url(request)
	var header_lines := _build_headers(headers)

	_apply_request_options(http, request)
	response._url = url
	response._method = request._method
	response._output_file = request._output_file
	if not request._output_file.is_empty():
		http.download_file = request._output_file

	var err := OK
	if request_body is PackedByteArray:
		err = http.request_raw(url, header_lines, request._method, request_body)
	else:
		err = http.request(url, header_lines, request._method, str(request_body))

	if err != OK:
		response._request_error = err
		response._error_message = "request start failed: %s" % error_string(err)
		http.queue_free()
		handle.call_deferred("_finish", response)
		return handle

	http.request_completed.connect(func(request_result: int, status_code: int, response_header_lines: PackedStringArray, response_body: PackedByteArray) -> void:
		http.queue_free()
		response._request_error = OK
		response._request_result = request_result
		response._status_code = status_code
		response._header_lines = response_header_lines
		response._headers = _parse_headers(response_header_lines)
		response._body = response_body

		if response._request_result != HTTPRequest.RESULT_SUCCESS:
			response._error_message = "request failed: %s" % _result_name(response._request_result)
		elif response._status_code < 200 or response._status_code >= 300:
			response._error_message = "http status: %d" % response._status_code

		if _should_parse_json(request):
			_parse_json_response(response)

		handle._finish(response)
	)

	return handle

static func _build_url(request: Variant) -> String:
	var endpoint: String = request._url
	for key in request._path_params:
		var value := str(request._path_params[key]).uri_encode()
		endpoint = endpoint.replace("{%s}" % str(key), value)

	var url := endpoint
	if not _is_absolute_url(url):
		url = _join_url(request._base_url, endpoint)

	var query := _encode_query(request._query_params)
	if query.is_empty():
		return url

	return "%s&%s" % [url, query] if url.contains("?") else "%s?%s" % [url, query]

static func _build_body(request: Variant, headers: Dictionary) -> Variant:
	match request._body_format:
		RestyRequest.BodyFormat.RAW:
			if not request._body_content_type.is_empty():
				_set_header(headers, "Content-Type", request._body_content_type)
			return request._body
		RestyRequest.BodyFormat.FORM:
			if not request._body_content_type.is_empty():
				_set_header(headers, "Content-Type", request._body_content_type)
			return _encode_query(request._body)
		RestyRequest.BodyFormat.JSON:
			if not request._body_content_type.is_empty():
				_set_header(headers, "Content-Type", request._body_content_type)
			return JSON.stringify(request._body)
		_:
			if request._body == null:
				return ""
			if request._body is String:
				_set_header_if_missing(headers, "Content-Type", "text/plain")
				return request._body
			if request._body is PackedByteArray:
				_set_header_if_missing(headers, "Content-Type", "application/octet-stream")
				return request._body
			if request._body is Dictionary or request._body is Array:
				_set_header_if_missing(headers, "Content-Type", "application/json")
				return JSON.stringify(request._body)
			_set_header_if_missing(headers, "Content-Type", "text/plain")
			return str(request._body)

static func _build_headers(headers: Dictionary) -> PackedStringArray:
	var header_lines := PackedStringArray()
	for key in headers:
		header_lines.append("%s: %s" % [str(key), str(headers[key])])
	return header_lines

static func _parse_headers(header_lines: PackedStringArray) -> Dictionary:
	var headers := {}
	for line in header_lines:
		var index := line.find(":")
		if index < 0:
			continue
		var name := line.substr(0, index).strip_edges()
		var value := line.substr(index + 1).strip_edges()
		_set_header(headers, name, value)
	return headers

static func _encode_query(values: Dictionary) -> String:
	var parts := PackedStringArray()
	for key in values:
		var value: Variant = values[key]
		if value == null:
			continue
		if value is Array:
			for item in value:
				if item != null:
					parts.append("%s=%s" % [str(key).uri_encode(), str(item).uri_encode()])
		else:
			parts.append("%s=%s" % [str(key).uri_encode(), str(value).uri_encode()])
	return "&".join(parts)

static func _should_parse_json(request: RestyRequest) -> bool:
	if not request._output_file.is_empty():
		return false
	return request._http_options.parse_json

static func _parse_json_response(response: RestyResponse) -> void:
	var content_type := response.get_header("Content-Type").to_lower()
	var text := response.text.strip_edges()
	if text.is_empty():
		return
	if not content_type.contains("json") and not text.begins_with("{") and not text.begins_with("["):
		return

	var parser := JSON.new()
	var err := parser.parse(text)
	if err != OK:
		response._json_error_message = parser.get_error_message()
		return
	response._json = parser.data

static func _apply_request_options(http: HTTPRequest, request: RestyRequest) -> void:
	http.timeout = request._http_options.timeout
	http.accept_gzip = request._http_options.accept_gzip
	http.body_size_limit = request._http_options.body_size_limit
	http.download_chunk_size = request._http_options.download_chunk_size
	http.max_redirects = request._http_options.max_redirects
	http.use_threads = request._http_options.use_threads

static func _set_header_if_missing(headers: Dictionary, name: String, value: String) -> void:
	var canonical_name := _canonical_header_name(name)
	if not headers.has(canonical_name):
		headers[canonical_name] = value

static func _set_header(headers: Dictionary, name: String, value: String) -> void:
	headers[_canonical_header_name(name)] = value

static func _has_header(headers: Dictionary, name: String) -> bool:
	return headers.has(_canonical_header_name(name))

static func _canonical_header_name(name: String) -> String:
	if _is_canonical_header_name(name):
		return name

	var parts := PackedStringArray()
	for part in name.strip_edges().split("-"):
		var value := String(part)
		if value.is_empty():
			parts.append(value)
			continue
		parts.append(value.substr(0, 1).to_upper() + value.substr(1).to_lower())
	return "-".join(parts)

static func _is_canonical_header_name(name: String) -> bool:
	if name.is_empty():
		return true

	var upper_next := true
	for index in range(name.length()):
		var code := name.unicode_at(index)
		if code <= 32:
			return false
		if code == 45:
			upper_next = true
			continue
		if code >= 65 and code <= 90:
			if not upper_next:
				return false
		elif code >= 97 and code <= 122:
			if upper_next:
				return false
		upper_next = false
	return true

static func _is_absolute_url(url: String) -> bool:
	return url.begins_with("http://") or url.begins_with("https://")

static func _join_url(prefix: String, endpoint: String) -> String:
	if prefix.is_empty():
		return endpoint
	if endpoint.is_empty():
		return prefix
	if prefix.ends_with("/") and endpoint.begins_with("/"):
		return prefix.substr(0, prefix.length() - 1) + endpoint
	if not prefix.ends_with("/") and not endpoint.begins_with("/"):
		return "%s/%s" % [prefix, endpoint]
	return prefix + endpoint

static func _result_name(value: int) -> String:
	match value:
		HTTPRequest.RESULT_SUCCESS:
			return "success"
		HTTPRequest.RESULT_CHUNKED_BODY_SIZE_MISMATCH:
			return "chunked body size mismatch"
		HTTPRequest.RESULT_CANT_CONNECT:
			return "can't connect"
		HTTPRequest.RESULT_CANT_RESOLVE:
			return "can't resolve"
		HTTPRequest.RESULT_CONNECTION_ERROR:
			return "connection error"
		HTTPRequest.RESULT_TLS_HANDSHAKE_ERROR:
			return "tls handshake error"
		HTTPRequest.RESULT_NO_RESPONSE:
			return "no response"
		HTTPRequest.RESULT_BODY_SIZE_LIMIT_EXCEEDED:
			return "body size limit exceeded"
		HTTPRequest.RESULT_BODY_DECOMPRESS_FAILED:
			return "body decompress failed"
		HTTPRequest.RESULT_REQUEST_FAILED:
			return "request failed"
		HTTPRequest.RESULT_DOWNLOAD_FILE_CANT_OPEN:
			return "download file can't open"
		HTTPRequest.RESULT_DOWNLOAD_FILE_WRITE_ERROR:
			return "download file write error"
		HTTPRequest.RESULT_REDIRECT_LIMIT_REACHED:
			return "redirect limit reached"
		HTTPRequest.RESULT_TIMEOUT:
			return "timeout"
		_:
			return "unknown result %d" % value
