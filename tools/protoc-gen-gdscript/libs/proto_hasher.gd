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
# Incremental writer contract used by ProtoUtils hash helpers.
@abstract
class_name ProtoHasher
extends RefCounted

@abstract
func write_byte(value: int) -> void

@abstract
func write_int32(value: int) -> void

@abstract
func write_uint32(value: int) -> void

@abstract
func write_int64(value: int) -> void

@abstract
func write_uint64(value: int) -> void

@abstract
func write_bytes(value: PackedByteArray) -> void

@abstract
func sum64() -> int
