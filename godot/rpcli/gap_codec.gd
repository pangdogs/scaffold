class_name GAPCodec
extends RefCounted

const ByteStream = preload("byte_stream.gd")

static func encode_packet(body: GAPMessages.Msg, src: GAPMessages.Origin = null, seq: int = 0) -> PackedByteArray:
	var head := GAPMessages.MsgHead.new()
	head.msg_id = body.msg_id()
	if src:
		head.src = src
	head.seq = seq
	head.length = head.size() + body.size()
	var stream := ByteStream.new()
	if !head.serialize(stream):
		push_error("failed to serialize GAP packet head, msg_id=%d, err=%s, message=%s" % [head.msg_id, stream.get_error(), stream.get_error_message()])
		return PackedByteArray()
	if !body.serialize(stream):
		push_error("failed to serialize GAP packet body, msg_id=%d, err=%s, message=%s" % [head.msg_id, stream.get_error(), stream.get_error_message()])
		return PackedByteArray()
	return stream.bytes()

static func decode_packet(data: PackedByteArray) -> GAPMessages.MsgPacket:
	var stream := ByteStream.new(data)
	var head := GAPMessages.MsgHead.new()
	if !head.deserialize(stream):
		push_error("failed to deserialize GAP packet head, err=%s, message=%s" % [stream.get_error(), stream.get_error_message()])
		return null
	if head.length != data.size():
		push_error("invalid GAP packet length, msg_id=%d" % head.msg_id)
		return null
	var body := GAPMessages.new_message(head.msg_id)
	if body == null:
		push_error("failed to new GAP packet body message, msg_id=%d" % head.msg_id)
		return null
	if !body.deserialize(stream):
		push_error("failed to deserialize GAP packet body, msg_id=%d, err=%s, message=%s" % [head.msg_id, stream.get_error(), stream.get_error_message()])
		return null
	return GAPMessages.MsgPacket.new(head, body)
