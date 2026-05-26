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
class_name RestyRequest
extends RefCounted

enum BodyFormat {
	AUTO,
	JSON,
	RAW,
	FORM,
}

var _client: RestyClient = null
var _base_url := ""
var _http_options: RestyHttpOptions = null
var _headers := {}
var _query_params := {}
var _path_params := {}
var _body: Variant = null
var _body_format := BodyFormat.AUTO
var _body_content_type := ""
var _method: int = HTTPClient.METHOD_GET
var _url: String = ""
var _output_file := ""

func _init(client: RestyClient, base_url: String, http_options: RestyHttpOptions, headers: Dictionary, query_params: Dictionary) -> void:
	_client = client
	_base_url = base_url
	_http_options = http_options.duplicate()
	_headers = headers.duplicate(true)
	_query_params = query_params.duplicate(true)

func set_header(name: String, value: Variant) -> RestyRequest:
	RestyClient._set_header(_headers, name, str(value))
	return self

func set_headers(values: Dictionary) -> RestyRequest:
	for key in values:
		set_header(str(key), values[key])
	return self

func set_timeout(seconds: float) -> RestyRequest:
	_http_options.timeout = seconds
	return self

func set_parse_json(enabled: bool) -> RestyRequest:
	_http_options.parse_json = enabled
	return self

func set_accept_gzip(enabled: bool) -> RestyRequest:
	_http_options.accept_gzip = enabled
	return self

func set_body_size_limit(bytes: int) -> RestyRequest:
	_http_options.body_size_limit = bytes
	return self

func set_download_chunk_size(bytes: int) -> RestyRequest:
	_http_options.download_chunk_size = bytes
	return self

func set_max_redirects(count: int) -> RestyRequest:
	_http_options.max_redirects = count
	return self

func set_use_threads(enabled: bool) -> RestyRequest:
	_http_options.use_threads = enabled
	return self

func set_output(path: String) -> RestyRequest:
	_output_file = path
	return self

func set_bearer_auth(token: String) -> RestyRequest:
	return set_header("Authorization", "Bearer %s" % token)

func set_basic_auth(username: String, password: String) -> RestyRequest:
	var token := Marshalls.utf8_to_base64("%s:%s" % [username, password])
	return set_header("Authorization", "Basic %s" % token)

func set_query_param(name: String, value: Variant) -> RestyRequest:
	_query_params[name] = value
	return self

func set_query_params(values: Dictionary) -> RestyRequest:
	for key in values:
		set_query_param(str(key), values[key])
	return self

func set_path_param(name: String, value: Variant) -> RestyRequest:
	_path_params[name] = value
	return self

func set_path_params(values: Dictionary) -> RestyRequest:
	for key in values:
		set_path_param(str(key), values[key])
	return self

func set_content_type(value: String) -> RestyRequest:
	return set_header("Content-Type", value)

func set_body(value: Variant) -> RestyRequest:
	_body = value
	_body_format = BodyFormat.AUTO
	_body_content_type = ""
	return self

func set_raw_body(value: PackedByteArray, content_type: String = "application/octet-stream") -> RestyRequest:
	_body = value
	_body_format = BodyFormat.RAW
	_body_content_type = content_type
	return self

func set_json(value: Variant, content_type: String = "application/json") -> RestyRequest:
	_body = value
	_body_format = BodyFormat.JSON
	_body_content_type = content_type
	return self

func set_form(values: Dictionary, content_type: String = "application/x-www-form-urlencoded") -> RestyRequest:
	_body = values
	_body_format = BodyFormat.FORM
	_body_content_type = content_type
	return self

func request_async(method: int, url: String) -> RestyResponse:
	_method = method
	_url = url
	return await _client._request_async(self)

func request_start(method: int, url: String) -> RestyRequestHandle:
	_method = method
	_url = url
	return _client._request_start(self)

func get_async(url: String) -> RestyResponse:
	return await request_async(HTTPClient.METHOD_GET, url)

func get_start(url: String) -> RestyRequestHandle:
	return request_start(HTTPClient.METHOD_GET, url)

func post_async(url: String) -> RestyResponse:
	return await request_async(HTTPClient.METHOD_POST, url)

func post_start(url: String) -> RestyRequestHandle:
	return request_start(HTTPClient.METHOD_POST, url)

func put_async(url: String) -> RestyResponse:
	return await request_async(HTTPClient.METHOD_PUT, url)

func put_start(url: String) -> RestyRequestHandle:
	return request_start(HTTPClient.METHOD_PUT, url)

func patch_async(url: String) -> RestyResponse:
	return await request_async(HTTPClient.METHOD_PATCH, url)

func patch_start(url: String) -> RestyRequestHandle:
	return request_start(HTTPClient.METHOD_PATCH, url)

func delete_async(url: String) -> RestyResponse:
	return await request_async(HTTPClient.METHOD_DELETE, url)

func delete_start(url: String) -> RestyRequestHandle:
	return request_start(HTTPClient.METHOD_DELETE, url)

func head_async(url: String) -> RestyResponse:
	return await request_async(HTTPClient.METHOD_HEAD, url)

func head_start(url: String) -> RestyRequestHandle:
	return request_start(HTTPClient.METHOD_HEAD, url)
