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
class_name RestyHttpOptions
extends RefCounted

var timeout: float = 30.0
var parse_json: bool = true
var accept_gzip: bool = true
var body_size_limit: int = -1
var download_chunk_size: int = 64 * 1024
var max_redirects: int = 8
var use_threads: bool = false

func duplicate() -> RestyHttpOptions:
	var copied := RestyHttpOptions.new()
	copied.timeout = timeout
	copied.parse_json = parse_json
	copied.accept_gzip = accept_gzip
	copied.body_size_limit = body_size_limit
	copied.download_chunk_size = download_chunk_size
	copied.max_redirects = max_redirects
	copied.use_threads = use_threads
	return copied
