class_name GTPMessages
extends RefCounted

const ByteStream = preload("byte_stream.gd")

const VERSION_V1_0 := 0x0100

const FLAG_ENCRYPTED := 1 << 0
const FLAG_SIGNED := 1 << 1
const FLAG_COMPRESSED := 1 << 2
const FLAG_CUSTOMIZE := 3

const FLAG_HELLO_DONE := 1 << FLAG_CUSTOMIZE
const FLAG_ENCRYPTION := 1 << (FLAG_CUSTOMIZE + 1)
const FLAG_AUTH := 1 << (FLAG_CUSTOMIZE + 2)
const FLAG_CONTINUE := 1 << (FLAG_CUSTOMIZE + 3)

const FLAG_SIGNATURE := 1 << FLAG_CUSTOMIZE
const FLAG_VERIFY_ENCRYPTION := 1 << FLAG_CUSTOMIZE
const FLAG_REQ_TIME := 1 << FLAG_CUSTOMIZE
const FLAG_RESP_TIME := 1 << (FLAG_CUSTOMIZE + 1)
const FLAG_PING := 1 << FLAG_CUSTOMIZE
const FLAG_PONG := 1 << (FLAG_CUSTOMIZE + 1)
const FLAG_ENCRYPT_OK := 1 << FLAG_CUSTOMIZE
const FLAG_AUTH_OK := 1 << (FLAG_CUSTOMIZE + 1)
const FLAG_CONTINUE_OK := 1 << (FLAG_CUSTOMIZE + 2)

const MSG_ID_NONE := 0
const MSG_ID_HELLO := 1
const MSG_ID_ECDHE_SECRET_KEY_EXCHANGE := 2
const MSG_ID_CHANGE_CIPHER_SPEC := 3
const MSG_ID_AUTH := 4
const MSG_ID_CONTINUE := 5
const MSG_ID_FINISHED := 6
const MSG_ID_RST := 7
const MSG_ID_HEARTBEAT := 8
const MSG_ID_SYNC_TIME := 9
const MSG_ID_PAYLOAD := 10

const SECRET_KEY_EXCHANGE_NONE := 0
const SECRET_KEY_EXCHANGE_ECDHE := 1

const SYMMETRIC_ENCRYPTION_NONE := 0
const SYMMETRIC_ENCRYPTION_AES := 1
const SYMMETRIC_ENCRYPTION_CHACHA20 := 2
const SYMMETRIC_ENCRYPTION_XCHACHA20 := 3
const SYMMETRIC_ENCRYPTION_CHACHA20_POLY1305 := 4
const SYMMETRIC_ENCRYPTION_XCHACHA20_POLY1305 := 5

const BLOCK_CIPHER_MODE_NONE := 0
const PADDING_MODE_NONE := 0
const HASH_NONE := 0

const COMPRESSION_NONE := 0
const COMPRESSION_GZIP := 1
const COMPRESSION_DEFLATE := 2
const COMPRESSION_BROTLI := 3
const COMPRESSION_LZ4 := 4
const COMPRESSION_SNAPPY := 5

const CODE_VERSION_ERROR := 1
const CODE_SESSION_NOT_FOUND := 2
const CODE_ENCRYPT_FAILED := 3
const CODE_AUTH_FAILED := 4
const CODE_CONTINUE_FAILED := 5
const CODE_REJECT := 6
const CODE_SHUTDOWN := 7
const CODE_SESSION_DEATH := 8

class CipherSuite:
	extends RefCounted

	var secret_key_exchange: int = SECRET_KEY_EXCHANGE_NONE
	var symmetric_encryption: int = SYMMETRIC_ENCRYPTION_NONE
	var block_cipher_mode: int = BLOCK_CIPHER_MODE_NONE
	var padding_mode: int = PADDING_MODE_NONE
	var hmac: int = HASH_NONE

	func serialize(stream: ByteStream) -> bool:
		stream.write_u8(secret_key_exchange)
		stream.write_u8(symmetric_encryption)
		stream.write_u8(block_cipher_mode)
		stream.write_u8(padding_mode)
		stream.write_u8(hmac)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		secret_key_exchange = stream.read_u8()
		symmetric_encryption = stream.read_u8()
		block_cipher_mode = stream.read_u8()
		padding_mode = stream.read_u8()
		hmac = stream.read_u8()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.SIZEOF_U8 * 5

class SignatureAlgorithm:
	extends RefCounted

	var asymmetric_encryption: int = 0
	var padding_mode: int = 0
	var hash: int = 0

	func serialize(stream: ByteStream) -> bool:
		stream.write_u8(asymmetric_encryption)
		stream.write_u8(padding_mode)
		stream.write_u8(hash)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		asymmetric_encryption = stream.read_u8()
		padding_mode = stream.read_u8()
		hash = stream.read_u8()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.SIZEOF_U8 * 3

class MsgHead:
	extends RefCounted

	var length: int = 0
	var msg_id: int = MSG_ID_NONE
	var flags: int = 0
	var seq: int = 0
	var ack: int = 0

	func serialize(stream: ByteStream) -> bool:
		stream.write_u32(length)
		stream.write_u8(msg_id)
		stream.write_u8(flags)
		stream.write_u32(seq)
		stream.write_u32(ack)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		length = stream.read_u32()
		msg_id = stream.read_u8()
		flags = stream.read_u8()
		seq = stream.read_u32()
		ack = stream.read_u32()
		return stream.get_error() == OK

	func size() -> int:
		return (
			ByteStream.SIZEOF_U32 +
			ByteStream.SIZEOF_U8 +
			ByteStream.SIZEOF_U8 +
			ByteStream.SIZEOF_U32 +
			ByteStream.SIZEOF_U32
		)

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

class MsgHello:
	extends Msg

	var version: int = VERSION_V1_0
	var session_id: String = ""
	var random: PackedByteArray = PackedByteArray()
	var cipher_suite: CipherSuite = CipherSuite.new()
	var compression: int = COMPRESSION_NONE

	func msg_id() -> int:
		return MSG_ID_HELLO

	func serialize(stream: ByteStream) -> bool:
		stream.write_u16(version)
		stream.write_string(session_id)
		stream.write_bytes(random)
		if cipher_suite == null or !cipher_suite.serialize(stream):
			return false
		stream.write_u8(compression)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		version = stream.read_u16()
		session_id = stream.read_string()
		random = stream.read_bytes()
		if cipher_suite == null or !cipher_suite.deserialize(stream):
			return false
		compression = stream.read_u8()
		return stream.get_error() == OK

	func size() -> int:
		return (
			ByteStream.SIZEOF_U16 +
			ByteStream.sizeof_string(session_id) +
			ByteStream.sizeof_bytes(random) +
			(cipher_suite.size() if cipher_suite != null else 0) +
			ByteStream.SIZEOF_U8
		)

class MsgECDHESecretKeyExchange:
	extends Msg

	var named_curve: int = 0
	var public_key: PackedByteArray = PackedByteArray()
	var iv: PackedByteArray = PackedByteArray()
	var nonce: PackedByteArray = PackedByteArray()
	var nonce_step: PackedByteArray = PackedByteArray()
	var signature_algorithm: SignatureAlgorithm = SignatureAlgorithm.new()
	var signature: PackedByteArray = PackedByteArray()

	func msg_id() -> int:
		return MSG_ID_ECDHE_SECRET_KEY_EXCHANGE

	func serialize(stream: ByteStream) -> bool:
		stream.write_u8(named_curve)
		stream.write_bytes(public_key)
		stream.write_bytes(iv)
		stream.write_bytes(nonce)
		stream.write_bytes(nonce_step)
		if signature_algorithm == null or !signature_algorithm.serialize(stream):
			return false
		stream.write_bytes(signature)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		named_curve = stream.read_u8()
		public_key = stream.read_bytes()
		iv = stream.read_bytes()
		nonce = stream.read_bytes()
		nonce_step = stream.read_bytes()
		if signature_algorithm == null or !signature_algorithm.deserialize(stream):
			return false
		signature = stream.read_bytes()
		return stream.get_error() == OK

	func size() -> int:
		return (
			ByteStream.SIZEOF_U8 +
			ByteStream.sizeof_bytes(public_key) +
			ByteStream.sizeof_bytes(iv) +
			ByteStream.sizeof_bytes(nonce) +
			ByteStream.sizeof_bytes(nonce_step) +
			(signature_algorithm.size() if signature_algorithm != null else 0) +
			ByteStream.sizeof_bytes(signature)
		)

class MsgChangeCipherSpec:
	extends Msg

	var encrypted_hello: PackedByteArray = PackedByteArray()

	func msg_id() -> int:
		return MSG_ID_CHANGE_CIPHER_SPEC

	func serialize(stream: ByteStream) -> bool:
		stream.write_bytes(encrypted_hello)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		encrypted_hello = stream.read_bytes()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.sizeof_bytes(encrypted_hello)

class MsgAuth:
	extends Msg

	var user_id: String = ""
	var token: String = ""
	var extensions: PackedByteArray = PackedByteArray()

	func msg_id() -> int:
		return MSG_ID_AUTH

	func serialize(stream: ByteStream) -> bool:
		stream.write_string(user_id)
		stream.write_string(token)
		stream.write_bytes(extensions)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		user_id = stream.read_string()
		token = stream.read_string()
		extensions = stream.read_bytes()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.sizeof_string(user_id) + ByteStream.sizeof_string(token) + ByteStream.sizeof_bytes(extensions)

class MsgContinue:
	extends Msg

	var send_seq: int = 0
	var recv_seq: int = 0

	func msg_id() -> int:
		return MSG_ID_CONTINUE

	func serialize(stream: ByteStream) -> bool:
		stream.write_u32(send_seq)
		stream.write_u32(recv_seq)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		send_seq = stream.read_u32()
		recv_seq = stream.read_u32()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.SIZEOF_U32 + ByteStream.SIZEOF_U32

class MsgFinished:
	extends Msg

	var send_seq: int = 0
	var recv_seq: int = 0

	func msg_id() -> int:
		return MSG_ID_FINISHED

	func serialize(stream: ByteStream) -> bool:
		stream.write_u32(send_seq)
		stream.write_u32(recv_seq)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		send_seq = stream.read_u32()
		recv_seq = stream.read_u32()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.SIZEOF_U32 + ByteStream.SIZEOF_U32

class MsgRst:
	extends Msg

	var code: int = 0
	var message: String = ""

	func msg_id() -> int:
		return MSG_ID_RST

	func serialize(stream: ByteStream) -> bool:
		stream.write_i32(code)
		stream.write_string(message)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		code = stream.read_i32()
		message = stream.read_string()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.SIZEOF_I32 + ByteStream.sizeof_string(message)

class MsgHeartbeat:
	extends Msg

	func msg_id() -> int:
		return MSG_ID_HEARTBEAT

	func serialize(stream: ByteStream) -> bool:
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		return stream.get_error() == OK

	func size() -> int:
		return 0

class MsgSyncTime:
	extends Msg

	var corr_id: int = 0
	var origin_time: int = 0
	var receive_time: int = 0
	var transmit_time: int = 0
	var zone_offset: int = 0

	func msg_id() -> int:
		return MSG_ID_SYNC_TIME

	func serialize(stream: ByteStream) -> bool:
		stream.write_i64(corr_id)
		stream.write_i64(origin_time)
		stream.write_i64(receive_time)
		stream.write_i64(transmit_time)
		stream.write_i32(zone_offset)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		corr_id = stream.read_i64()
		origin_time = stream.read_i64()
		receive_time = stream.read_i64()
		transmit_time = stream.read_i64()
		zone_offset = stream.read_i32()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.SIZEOF_I64 + ByteStream.SIZEOF_I64 + ByteStream.SIZEOF_I64 + ByteStream.SIZEOF_I64 + ByteStream.SIZEOF_I32

class MsgPayload:
	extends Msg

	var data: PackedByteArray = PackedByteArray()

	func msg_id() -> int:
		return MSG_ID_PAYLOAD

	func serialize(stream: ByteStream) -> bool:
		stream.write_bytes(data)
		return stream.get_error() == OK

	func deserialize(stream: ByteStream) -> bool:
		data = stream.read_bytes()
		return stream.get_error() == OK

	func size() -> int:
		return ByteStream.sizeof_bytes(data)

static func new_message(msg_id: int) -> Msg:
	match msg_id:
		MSG_ID_HELLO:
			return MsgHello.new()
		MSG_ID_ECDHE_SECRET_KEY_EXCHANGE:
			return MsgECDHESecretKeyExchange.new()
		MSG_ID_CHANGE_CIPHER_SPEC:
			return MsgChangeCipherSpec.new()
		MSG_ID_AUTH:
			return MsgAuth.new()
		MSG_ID_CONTINUE:
			return MsgContinue.new()
		MSG_ID_FINISHED:
			return MsgFinished.new()
		MSG_ID_RST:
			return MsgRst.new()
		MSG_ID_HEARTBEAT:
			return MsgHeartbeat.new()
		MSG_ID_SYNC_TIME:
			return MsgSyncTime.new()
		MSG_ID_PAYLOAD:
			return MsgPayload.new()
	return null
