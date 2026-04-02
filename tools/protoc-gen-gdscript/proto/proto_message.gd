# Base contract implemented by all generated protobuf message classes.
# Generated messages are expected to support serialization, deserialization,
# deep copying, and structural equality checks.
@abstract
class_name ProtoMessage
extends RefCounted

@abstract
# Writes the current message into the provided protobuf output stream.
func serialize(stream: ProtoOutputStream) -> bool

@abstract
# Populates the current message by consuming protobuf data from the stream.
func deserialize(stream: ProtoInputStream) -> bool

@abstract
# Returns the encoded protobuf payload size in bytes.
func size() -> int

@abstract
# Restores all fields to their generated default values.
func reset() -> void

@abstract
# Allocates a new empty message instance of the same concrete type.
func new() -> ProtoMessage

@abstract
# Creates a deep copy of the current message instance.
func clone() -> ProtoMessage

@abstract
# Writes the current message into the provided hash state.
func hash_to(hasher) -> void

# Returns the structural hash of the current message.
func hash() -> int:
	var hasher := ProtoUtils.new_hasher()
	ProtoUtils.hash_message(hasher, self)
	return hasher.sum64()

@abstract
# Compares two messages by value rather than by reference identity.
func equals(other: ProtoMessage) -> bool
