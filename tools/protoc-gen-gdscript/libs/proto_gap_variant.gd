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
# Base contract for generated protobuf messages that can be passed through
# Golaxy GAP custom variant encoding.
# Subclasses provide a stable custom variant type id derived from the protobuf
# package name and message name so both Go and GDScript runtimes can agree on
# the wire type.
@abstract
class_name ProtoGAPVariant
extends ProtoMessage

@abstract
# Returns the Golaxy GAP custom variant type id for this message type.
func type_id() -> int
