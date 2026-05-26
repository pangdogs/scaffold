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
class_name RestyResponse
extends RefCounted

var _request_error: int = OK
var _request_result: int = HTTPRequest.RESULT_SUCCESS
var _status_code: int = 0
var _headers := {}
var _header_lines: PackedStringArray = PackedStringArray()
var _body: PackedByteArray = PackedByteArray()
var _url: String = ""
var _method: int = HTTPClient.METHOD_GET
var _output_file: String = ""
var _error_message: String = ""
var _json: Variant = null
var _json_error_message: String = ""

var request_error: int:
	get:
		return _request_error

var request_result: int:
	get:
		return _request_result

var status_code: int:
	get:
		return _status_code

var headers: Dictionary:
	get:
		return _headers

var header_lines: PackedStringArray:
	get:
		return _header_lines

var body: PackedByteArray:
	get:
		return _body

var url: String:
	get:
		return _url

var method: int:
	get:
		return _method

var output_file: String:
	get:
		return _output_file

var error_message: String:
	get:
		return _error_message

var json: Variant:
	get:
		return _json

var json_error_message: String:
	get:
		return _json_error_message

var text: String:
	get:
		return _body.get_string_from_utf8()

func is_success() -> bool:
	return _request_error == OK and _request_result == HTTPRequest.RESULT_SUCCESS and _status_code >= 200 and _status_code < 300

func get_header(name: String) -> String:
	return str(_headers.get(RestyClient._canonical_header_name(name), ""))
