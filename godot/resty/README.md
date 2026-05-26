# Resty for Godot

Lightweight Resty-style HTTP helper for Godot 4.

## Autoload

Add `res://addons/resty/resty_client.gd` as an Autoload named `Resty`.

This project already has it registered in `project.godot`.

## Basic Usage

```gdscript
func _ready() -> void:
	(
		Resty.set_base_url("https://api.example.com")
		.set_header("Accept", "application/json")
	)

	var res := await (
		Resty.r()
		.set_bearer_auth("token")
		.set_query_param("page", 1)
		.get_async("/users")
	)

	if res.is_success():
		print(res.status_code)
		print(res.json)
	else:
		push_error(res.error_message)
```

## POST JSON

```gdscript
var res := await (
	Resty.r()
	.set_json({
		"name": "alice",
		"score": 100,
	})
	.post_async("/scores")
)
```

`set_body()` infers the request body format from the value type: `Dictionary` and `Array` are sent as JSON, `String` is sent as text, and `PackedByteArray` is sent as raw bytes.

If you already have a JSON string, send it as a plain body and set the content type explicitly:

```gdscript
var res := await (
	Resty.r()
	.set_body('{"name":"alice","score":100}')
	.set_content_type("application/json")
	.post_async("/scores")
)
```

## Raw Body

```gdscript
var res := await (
	Resty.r()
	.set_raw_body("abc".to_utf8_buffer())
	.post_async("/upload")
)
```

## POST Form

```gdscript
var res := await (
	Resty.r()
	.set_form({
		"name": "alice",
		"score": 100,
	})
	.post_async("/scores")
)
```

## Path Params

```gdscript
var res := await (
	Resty.r()
	.set_path_param("id", 42)
	.get_async("/users/{id}")
)
```

## Request Options

```gdscript
(
	Resty.set_timeout(10.0)
	.set_parse_json(true)
	.set_accept_gzip(true)
	.set_body_size_limit(-1)
	.set_download_chunk_size(65536)
	.set_max_redirects(8)
	.set_use_threads(false)
)

var res := await (
	Resty.r()
	.set_timeout(2.0)
	.set_parse_json(false)
	.set_use_threads(true)
	.get_async("/users")
)
```

`Resty.r()` snapshots the current base URL, default headers, default query params, and HTTP options. Changing `Resty` after creating a request does not affect that request.

## Concurrent Requests

Every request creates its own `HTTPRequest` node, so concurrent calls are supported:

```gdscript
var user_handle := Resty.r().get_start("/user")
var item_handle := Resty.r().get_start("/items")

var user: RestyResponse = await user_handle.completed
var items: RestyResponse = await item_handle.completed
```

## Server-Sent Events

```gdscript
var stream := (
	Resty.sse("/chat/stream")
	.set_bearer_auth("token")
)

stream.event_received.connect(func(event: RestySSEEvent) -> void:
	if event == null:
		return
	print(event.data)
)
stream.closed.connect(func(error_message: String) -> void:
	if not error_message.is_empty():
		push_error(error_message)
)

if not stream.start():
	push_error(stream.error_message)
```

`event_received` emits `null` once when the stream closes, so `await` loops can exit:

```gdscript
while true:
	var event: RestySSEEvent = await stream.event_received
	if event == null:
		break
	print(event.data)
```

SSE uses `GET` and automatically adds `Accept: text/event-stream` and `Cache-Control: no-cache` if they are not already set. The returned `RestySSEStream` is a long-lived stream; call `close()` to stop it.

Like `Resty.r()`, `Resty.sse()` snapshots the current client settings when the stream is created.

## Download To File

```gdscript
var res := await (
	Resty.r()
	.set_output("user://patch.zip")
	.get_async("https://example.com/patch.zip")
)

if not res.is_success():
	push_error(res.error_message)
```

## Response

```gdscript
res.request_error
res.request_result
res.status_code
res.headers
res.header_lines
res.body
res.text
res.url
res.method
res.json
res.output_file
res.error_message
res.json_error_message
res.is_success()
res.get_header("Content-Type")
```

## Client API

```gdscript
Resty.r() -> RestyRequest
Resty.sse(url: String) -> RestySSEStream

Resty.set_base_url(value: String) -> RestyClient
Resty.set_header(name: String, value: Variant) -> RestyClient
Resty.set_headers(values: Dictionary) -> RestyClient
Resty.set_timeout(seconds: float) -> RestyClient
Resty.set_parse_json(enabled: bool) -> RestyClient
Resty.set_accept_gzip(enabled: bool) -> RestyClient
Resty.set_body_size_limit(bytes: int) -> RestyClient
Resty.set_download_chunk_size(bytes: int) -> RestyClient
Resty.set_max_redirects(count: int) -> RestyClient
Resty.set_use_threads(enabled: bool) -> RestyClient
Resty.set_bearer_auth(token: String) -> RestyClient
Resty.set_basic_auth(username: String, password: String) -> RestyClient
Resty.set_query_param(name: String, value: Variant) -> RestyClient
Resty.set_query_params(values: Dictionary) -> RestyClient

Resty.get_async(url: String) -> RestyResponse
Resty.get_start(url: String) -> RestyRequestHandle
Resty.post_async(url: String, body: Variant = null) -> RestyResponse
Resty.post_start(url: String, body: Variant = null) -> RestyRequestHandle
Resty.put_async(url: String, body: Variant = null) -> RestyResponse
Resty.put_start(url: String, body: Variant = null) -> RestyRequestHandle
Resty.patch_async(url: String, body: Variant = null) -> RestyResponse
Resty.patch_start(url: String, body: Variant = null) -> RestyRequestHandle
Resty.delete_async(url: String) -> RestyResponse
Resty.delete_start(url: String) -> RestyRequestHandle
Resty.head_async(url: String) -> RestyResponse
Resty.head_start(url: String) -> RestyRequestHandle
```

## Request API

```gdscript
req.set_header(name: String, value: Variant) -> RestyRequest
req.set_headers(values: Dictionary) -> RestyRequest
req.set_timeout(seconds: float) -> RestyRequest
req.set_parse_json(enabled: bool) -> RestyRequest
req.set_accept_gzip(enabled: bool) -> RestyRequest
req.set_body_size_limit(bytes: int) -> RestyRequest
req.set_download_chunk_size(bytes: int) -> RestyRequest
req.set_max_redirects(count: int) -> RestyRequest
req.set_use_threads(enabled: bool) -> RestyRequest
req.set_output(path: String) -> RestyRequest
req.set_bearer_auth(token: String) -> RestyRequest
req.set_basic_auth(username: String, password: String) -> RestyRequest
req.set_query_param(name: String, value: Variant) -> RestyRequest
req.set_query_params(values: Dictionary) -> RestyRequest
req.set_path_param(name: String, value: Variant) -> RestyRequest
req.set_path_params(values: Dictionary) -> RestyRequest
req.set_content_type(value: String) -> RestyRequest
req.set_body(value: Variant) -> RestyRequest
req.set_raw_body(value: PackedByteArray, content_type: String = "application/octet-stream") -> RestyRequest
req.set_json(value: Variant, content_type: String = "application/json") -> RestyRequest
req.set_form(values: Dictionary, content_type: String = "application/x-www-form-urlencoded") -> RestyRequest

req.request_async(method: int, url: String) -> RestyResponse
req.request_start(method: int, url: String) -> RestyRequestHandle

req.get_async(url: String) -> RestyResponse
req.get_start(url: String) -> RestyRequestHandle
req.post_async(url: String) -> RestyResponse
req.post_start(url: String) -> RestyRequestHandle
req.put_async(url: String) -> RestyResponse
req.put_start(url: String) -> RestyRequestHandle
req.patch_async(url: String) -> RestyResponse
req.patch_start(url: String) -> RestyRequestHandle
req.delete_async(url: String) -> RestyResponse
req.delete_start(url: String) -> RestyRequestHandle
req.head_async(url: String) -> RestyResponse
req.head_start(url: String) -> RestyRequestHandle
```

## SSE API

```gdscript
signal opened()
signal event_received(event: RestySSEEvent)
signal closed(error_message: String)

stream.set_header(name: String, value: Variant) -> RestySSEStream
stream.set_headers(values: Dictionary) -> RestySSEStream
stream.set_timeout(seconds: float) -> RestySSEStream
stream.set_bearer_auth(token: String) -> RestySSEStream
stream.set_basic_auth(username: String, password: String) -> RestySSEStream
stream.set_query_param(name: String, value: Variant) -> RestySSEStream
stream.set_query_params(values: Dictionary) -> RestySSEStream
stream.set_path_param(name: String, value: Variant) -> RestySSEStream
stream.set_path_params(values: Dictionary) -> RestySSEStream

stream.start() -> bool
stream.close() -> void
stream.get_response_header(name: String) -> String

stream.status_code
stream.response_headers
stream.response_header_lines
stream.last_event_id
stream.error_message
```
