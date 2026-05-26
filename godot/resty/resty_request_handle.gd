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
class_name RestyRequestHandle
extends RefCounted

signal completed(response: RestyResponse)

var _response: RestyResponse = null
var _done := false

var response: RestyResponse:
	get:
		return _response

var done: bool:
	get:
		return _done

func _finish(value: RestyResponse) -> void:
	_response = value
	_done = true
	completed.emit(value)
