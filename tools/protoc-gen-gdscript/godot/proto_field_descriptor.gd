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
# Shared protobuf metadata constants used by the generated GDScript runtime.
class_name ProtoFieldDescriptor
extends RefCounted

# Canonical protobuf field type identifiers.
enum FieldType {
	TYPE_DOUBLE = 1,        # double
	TYPE_FLOAT = 2,         # float
	TYPE_INT64 = 3,         # int64
	TYPE_UINT64 = 4,        # uint64
	TYPE_INT32 = 5,         # int32
	TYPE_FIXED64 = 6,       # fixed64
	TYPE_FIXED32 = 7,       # fixed32
	TYPE_BOOL = 8,          # bool
	TYPE_STRING = 9,        # string
	TYPE_GROUP = 10,        # group (deprecated)
	TYPE_MESSAGE = 11,      # message
	TYPE_BYTES = 12,        # bytes
	TYPE_UINT32 = 13,      # uint32
	TYPE_ENUM = 14,        # enum
	TYPE_SFIXED32 = 15,    # sfixed32
	TYPE_SFIXED64 = 16,    # sfixed64
	TYPE_SINT32 = 17,      # sint32
	TYPE_SINT64 = 18,      # sint64
	TYPE_MAP = 19,         # map
}

# Low-level protobuf wire encodings.
enum WireType {
	WIRETYPE_VARINT = 0,           # int32, int64, uint32, uint64, sint32, sint64, bool, enum
	WIRETYPE_FIXED64 = 1,          # fixed64, sfixed64, double
	WIRETYPE_LENGTH_DELIMITED = 2,  # string, bytes, embedded messages, packed repeated fields
	WIRETYPE_START_GROUP = 3,       # groups (deprecated)
	WIRETYPE_END_GROUP = 4,         # groups (deprecated)
	WIRETYPE_FIXED32 = 5,          # fixed32, sfixed32, float
}

# Protobuf field cardinality markers.
enum FieldLabel {
	LABEL_OPTIONAL = 1,  # optional
	LABEL_REQUIRED = 2,  # required (proto2 only, deprecated)
	LABEL_REPEATED = 3,  # repeated
}

# Maps a protobuf field type identifier to its wire encoding.
# Returns -1 when the type does not have a supported wire representation.
static func get_field_wire_type(field_type: int) -> int:
	match field_type:
		FieldType.TYPE_INT32, FieldType.TYPE_INT64, FieldType.TYPE_UINT32, FieldType.TYPE_UINT64, FieldType.TYPE_SINT32, FieldType.TYPE_SINT64, FieldType.TYPE_BOOL, FieldType.TYPE_ENUM:
			return WireType.WIRETYPE_VARINT
		FieldType.TYPE_FIXED64, FieldType.TYPE_SFIXED64, FieldType.TYPE_DOUBLE:
			return WireType.WIRETYPE_FIXED64
		FieldType.TYPE_STRING, FieldType.TYPE_BYTES, FieldType.TYPE_MESSAGE, FieldType.TYPE_MAP:
			return WireType.WIRETYPE_LENGTH_DELIMITED
		FieldType.TYPE_FIXED32, FieldType.TYPE_SFIXED32, FieldType.TYPE_FLOAT:
			return WireType.WIRETYPE_FIXED32
		_:
			return -1