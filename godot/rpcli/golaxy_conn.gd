@abstract
class_name GolaxyConn
extends RefCounted

const CONNECTING := 0
const CONNECTED := 1
const CLOSED := 2

@abstract
func poll() -> void

@abstract
func get_status() -> int

@abstract
func send_bytes(data: PackedByteArray) -> int

@abstract
func send_partial_bytes(data: PackedByteArray) -> int

@abstract
func read_bytes() -> PackedByteArray

@abstract
func has_pending_bytes() -> bool

@abstract
func close() -> void

@abstract
func get_error() -> int

@abstract
func get_error_message() -> String
