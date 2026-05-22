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
class_name GTPCodec
extends RefCounted

const ByteStream = preload("byte_stream.gd")

static func supports_hello(hello: GTPMessages.MsgHello) -> String:
	if hello.version != GTPMessages.VERSION_V1_0:
		return "unsupported GTP version %d" % hello.version
	if (
		hello.cipher_suite.secret_key_exchange != GTPMessages.SECRET_KEY_EXCHANGE_NONE or
		hello.cipher_suite.symmetric_encryption != GTPMessages.SYMMETRIC_ENCRYPTION_NONE or
		hello.cipher_suite.block_cipher_mode != GTPMessages.BLOCK_CIPHER_MODE_NONE or
		hello.cipher_suite.padding_mode != GTPMessages.PADDING_MODE_NONE or
		hello.cipher_suite.hmac != GTPMessages.HASH_NONE
	):
		return "unsupported GTP encrypted payloads"
	if hello.compression != GTPMessages.COMPRESSION_NONE:
		return "unsupported GTP compressed payloads"
	return ""

static func encode_packet(body: GTPMessages.Msg, flags: int = 0, seq: int = 0, ack: int = 0) -> PackedByteArray:
	var head := GTPMessages.MsgHead.new()
	head.length = head.size() + body.size()
	head.msg_id = body.msg_id()
	head.flags = flags & ~(GTPMessages.FLAG_ENCRYPTED | GTPMessages.FLAG_SIGNED | GTPMessages.FLAG_COMPRESSED)
	head.seq = seq
	head.ack = ack
	var stream := ByteStream.new()
	if !head.serialize(stream):
		push_error("failed to serialize GTP packet head, msg_id=%d, err=%s, message=%s" % [head.msg_id, stream.get_error(), stream.get_error_message()])
		return PackedByteArray()
	if !body.serialize(stream):
		push_error("failed to serialize GTP packet body, msg_id=%d, err=%s, message=%s" % [head.msg_id, stream.get_error(), stream.get_error_message()])
		return PackedByteArray()
	return stream.bytes()

static func peek_packet_length(data: PackedByteArray) -> int:
	if data.size() < ByteStream.SIZEOF_U32:
		return 0
	var stream := ByteStream.new(data)
	var length := stream.read_u32()
	if stream.get_error() != OK or length < ByteStream.SIZEOF_U32:
		push_error("invalid GTP packet length, length=%d" % length)
		return -1
	return length

static func decode_packet(data: PackedByteArray) -> GTPMessages.MsgPacket:
	var stream := ByteStream.new(data)
	var head := GTPMessages.MsgHead.new()
	if !head.deserialize(stream):
		push_error("failed to deserialize GTP packet head, err=%s, message=%s" % [stream.get_error(), stream.get_error_message()])
		return null
	if head.length != data.size():
		push_error("invalid GTP packet length, msg_id=%d" % head.msg_id)
		return null
	if (head.flags & GTPMessages.FLAG_ENCRYPTED) != 0 or (head.flags & GTPMessages.FLAG_SIGNED) != 0:
		push_error("unsupported GTP encrypted payloads, msg_id=%d" % head.msg_id)
		return null
	if (head.flags & GTPMessages.FLAG_COMPRESSED) != 0:
		push_error("unsupported GTP compressed payloads, msg_id=%d" % head.msg_id)
		return null
	var body := GTPMessages.new_message(head.msg_id)
	if body == null:
		push_error("failed to new GTP packet body message, msg_id=%d" % head.msg_id)
		return null
	if !body.deserialize(stream):
		push_error("failed to deserialize GTP packet body, msg_id=%d" % head.msg_id)
		return null
	return GTPMessages.MsgPacket.new(head, body)
