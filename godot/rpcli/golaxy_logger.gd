class_name GolaxyLogger
extends RefCounted

class Sink:
	static func print_line(message: String) -> void:
		print(message)

	static func print_rich_line(message: String) -> void:
		print_rich(message)

	static func print_warning(message: String) -> void:
		push_warning(message)

	static func print_error(message: String) -> void:
		push_error(message)

	static func print_call_stack() -> void:
		print_stack()

var _name := ""
var _instance_id: int = 0
var _debug_enabled := false
var _parent: GolaxyLogger = null

var debug_enabled: bool:
	set(b):
		if _parent == null:
			_debug_enabled = b
	get:
		return _parent.debug_enabled if _parent != null else _debug_enabled

func _init(name: String, instance_id: int = 0, debug_enabled: bool = false) -> void:
	_name = name
	_instance_id = instance_id
	_debug_enabled = debug_enabled

func named(name: String, instance_id: int = 0) -> GolaxyLogger:
	var logger : GolaxyLogger = get_script().new(name, instance_id)
	logger._parent = self
	return logger

func format(message: String, args: Array = []) -> String:
	return _prefix() + (message % args if not args.is_empty() else message)

func print(message: String, args: Array = []) -> void:
	Sink.print_line(format(message, args))

func print_rich(message: String, args: Array = []) -> void:
	Sink.print_rich_line(format(message, args))

func print_stack(message: String = "", args: Array = []) -> void:
	if not message.is_empty():
		print(message, args)
	Sink.print_call_stack()

func debug(message: String, args: Array = []) -> void:
	if not debug_enabled:
		return
	self.print(message, args)

func debug_rich(message: String, args: Array = []) -> void:
	if not debug_enabled:
		return
	self.print_rich(message, args)

func debug_stack(message: String = "", args: Array = []) -> void:
	if not debug_enabled:
		return
	self.print_stack(message, args)

func warning(message: String, args: Array = []) -> void:
	Sink.print_warning(format(message, args))

func error(message: String, args: Array = []) -> void:
	Sink.print_error(format(message, args))

func _prefix() -> String:
	return "[%s][%s][%d] " % [_timestamp(), _name, _instance_id]

static func _timestamp() -> String:
	var dt := Time.get_datetime_dict_from_system()
	return "%04d-%02d-%02d %02d:%02d:%02d.%03d" % [
		int(dt["year"]),
		int(dt["month"]),
		int(dt["day"]),
		int(dt["hour"]),
		int(dt["minute"]),
		int(dt["second"]),
		Time.get_ticks_msec() % 1000,
	]
