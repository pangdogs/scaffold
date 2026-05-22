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
class_name GAPMessages
extends RefCounted

const ByteStream = preload("byte_stream.gd")

const MSG_ID_NONE := 0
const MSG_ID_RPC_REQUEST := 1
const MSG_ID_RPC_REPLY := 2
const MSG_ID_ONEWAY_RPC := 3
const MSG_ID_FORWARD := 4

class Origin:
	extends RefCounted

	var svc: String = ""
	var addr: String = ""
	var timestamp: int = 0

	func serialize(stream: ByteStream) -> bool:
		stream.write_string(svc)
		stream.write_string(addr)
		stream.write_i64(timestamp)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		svc = stream.read_string()
		addr = stream.read_string()
		timestamp = stream.read_i64()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.sizeof_string(svc) + ByteStream.sizeof_string(addr) + ByteStream.SIZEOF_I64

class MsgHead:
	extends RefCounted

	var length: int = 0
	var msg_id: int = MSG_ID_NONE
	var src: Origin = Origin.new()
	var seq: int = 0

	func serialize(stream: ByteStream) -> bool:
		stream.write_u32(length)
		stream.write_u32(msg_id)
		if src == null or !src.serialize(stream):
			return false
		stream.write_i64(seq)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		length = stream.read_u32()
		msg_id = stream.read_u32()
		if src == null or !src.deserialize(stream):
			return false
		seq = stream.read_i64()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.SIZEOF_U32 + ByteStream.SIZEOF_U32 + src.size() + ByteStream.SIZEOF_I64

class MsgPacket:
	extends RefCounted

	var head: MsgHead = null
	var body: Msg = null

	func _init(head: MsgHead = null, body: Msg = null):
		self.head = head
		self.body = body

	func size() -> int:
		return (head.size() if head != null else 0) + (body.size() if body != null else 0)

@abstract
class Msg:
	extends RefCounted

	@abstract
	func msg_id() -> int

	@abstract
	func serialize(stream: ByteStream) -> bool

	@abstract
	func deserialize(stream: ByteStream) -> bool

	@abstract
	func size() -> int

class MsgRPCRequest:
	extends Msg

	var corr_id: int = 0
	var call_chain: GAPVariants.GAPCallChain = GAPVariants.GAPCallChain.new()
	var path := PackedByteArray()
	var args: GAPVariants.GAPArray = GAPVariants.GAPArray.new()

	func msg_id() -> int:
		return MSG_ID_RPC_REQUEST

	func serialize(stream: ByteStream) -> bool:
		stream.write_varint(corr_id)
		if call_chain == null or !call_chain.serialize(stream):
			return false
		stream.write_bytes(path)
		if args == null or !args.serialize(stream):
			return false
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		corr_id = stream.read_varint()
		if call_chain == null or !call_chain.deserialize(stream):
			return false
		path = stream.read_bytes()
		if args == null or !args.deserialize(stream):
			return false
		return stream.get_error() == OK

	func size() -> int:
		return (
			ByteStream.sizeof_varint(corr_id) +
			(call_chain.size() if call_chain != null else 0) +
			ByteStream.sizeof_bytes(path) +
			(args.size() if args != null else 0)
		)

class MsgRPCReply:
	extends Msg

	var corr_id: int = 0
	var rets: GAPVariants.GAPArray = GAPVariants.GAPArray.new()
	var error: GAPVariants.GAPError = GAPVariants.GAPError.new()

	func msg_id() -> int:
		return MSG_ID_RPC_REPLY

	func serialize(stream: ByteStream) -> bool:
		stream.write_varint(corr_id)
		if rets == null or !rets.serialize(stream):
			return false
		if error == null or !error.serialize(stream):
			return false
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		corr_id = stream.read_varint()
		if rets == null or !rets.deserialize(stream):
			return false
		if error == null or !error.deserialize(stream):
			return false
		return stream.get_error() == OK

	func size() -> int:
		return (
			ByteStream.sizeof_varint(corr_id) +
			(rets.size() if rets != null else 0) +
			(error.size() if error != null else 0)
		)

class MsgOnewayRPC:
	extends Msg

	var call_chain: GAPVariants.GAPCallChain = GAPVariants.GAPCallChain.new()
	var path: PackedByteArray = PackedByteArray()
	var args: GAPVariants.GAPArray = GAPVariants.GAPArray.new()

	func msg_id() -> int:
		return MSG_ID_ONEWAY_RPC

	func serialize(stream: ByteStream) -> bool:
		if call_chain == null or !call_chain.serialize(stream):
			return false
		stream.write_bytes(path)
		if args == null or !args.serialize(stream):
			return false
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		if call_chain == null or !call_chain.deserialize(stream):
			return false
		path = stream.read_bytes()
		if args == null or !args.deserialize(stream):
			return false
		return stream.get_error() == OK

	func size() -> int:
		return (
			(call_chain.size() if call_chain != null else 0) +
			ByteStream.sizeof_bytes(path) +
			(args.size() if args != null else 0)
		)

class MsgForward:
	extends Msg

	var src: Origin = Origin.new()
	var dst: String = ""
	var corr_id: int = 0
	var trans_id: int = MSG_ID_NONE
	var trans_data: PackedByteArray = PackedByteArray()

	func msg_id() -> int:
		return MSG_ID_FORWARD

	func serialize(stream: ByteStream) -> bool:
		if src == null or !src.serialize(stream):
			return false
		stream.write_string(dst)
		stream.write_varint(corr_id)
		stream.write_u32(trans_id)
		stream.write_bytes(trans_data)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		if src == null or !src.deserialize(stream):
			return false
		dst = stream.read_string()
		corr_id = stream.read_varint()
		trans_id = stream.read_u32()
		trans_data = stream.read_bytes()
		return stream.get_error() == OK

	func size() -> int:
		return (
			(src.size() if src != null else 0) +
			ByteStream.sizeof_string(dst) +
			ByteStream.sizeof_varint(corr_id) +
			ByteStream.SIZEOF_U32 +
			ByteStream.sizeof_bytes(trans_data)
		)

static func new_message(msg_id: int) -> Msg:
	match msg_id:
		MSG_ID_RPC_REQUEST:
			return MsgRPCRequest.new()
		MSG_ID_RPC_REPLY:
			return MsgRPCReply.new()
		MSG_ID_ONEWAY_RPC:
			return MsgOnewayRPC.new()
		MSG_ID_FORWARD:
			return MsgForward.new()
	return null
