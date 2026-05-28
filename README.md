# Scaffold
[English](./README.md) | [简体中文](./README.zh_CN.md)

## Overview
`scaffold` is the tooling and add-in companion to the [**Golaxy Distributed Service Development Framework**](https://github.com/pangdogs/framework). It fills the engineering pieces around the main runtime for Go services, Godot clients, and Excel table pipelines: protocol code generation, table data export, script hotfixing, and entity property synchronization.

This repository is not a complete application framework. It packages the build-time and runtime helpers commonly needed by Golaxy projects into reusable components:

- `addins`: service/runtime extensions that plug into `git.golaxy.org/framework`.
- `tools`: command-line generators and Protobuf plugins used during build-time generation.
- `godot`: runtime scripts that can be copied into a Godot project.

## Capabilities
| Module | Responsibility |
| --- | --- |
| `addins/goscr` | Service-level Go scripting add-in built on Yaegi for script-backed entities/components, script project loading, and local or remote hot reloading. |
| `addins/propview` | Runtime property view add-in for managed entity properties, load/save flows, revision advancement, and service/client synchronization. |
| `tools/propc` | Property-sync generator that scans annotated Go declarations and emits `propview`-based `*.sync.gen.go` wrappers. |
| `tools/excelc` | Excel table toolchain that turns `.xlsx` workbooks into table proto schemas, access code, and binary/JSON data files. |
| `tools/protoc-gen-go-excel` | Go Protobuf plugin that adds table lookup, index access, and loading helpers for Excel-generated structs. |
| `tools/protoc-gen-go-structure` | Go Protobuf plugin that adds deep-clone helpers for messages, slices, maps, bytes, and nested messages. |
| `tools/protoc-gen-go-variant` | Go Protobuf plugin that makes generated messages usable as GAP variant values in the Golaxy RPC stack. |
| `tools/protoc-gen-gdscript` | Protobuf plugin that emits Godot-facing GDScript message types, serialization, and deserialization logic. |
| `tools/protoc-gen-gdscript-excel` | Protobuf plugin that emits GDScript table wrappers and index lookup helpers for Excel-generated schemas. |
| `godot/rpcli` | Godot-side Golaxy RPC client runtime for GAP/GTP connections, reconnects, RPC calls, callbacks, and GAP variant transport. |
| `godot/resty` | Godot-side Resty-style HTTP runtime for fluent HTTP requests, JSON/form/raw bodies, downloads, concurrent requests, and Server-Sent Events. |

## Installation
Install the module itself when importing the add-in packages:

```bash
go get git.golaxy.org/scaffold@latest
```

Install only the command-line tools you need:

```bash
go install git.golaxy.org/scaffold/tools/excelc@latest
go install git.golaxy.org/scaffold/tools/propc@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-excel@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-structure@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-variant@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-gdscript@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-gdscript-excel@latest
```

When using the Protobuf plugins, make sure `protoc` is available and the installed `protoc-gen-*` binaries are on `PATH`. This module currently declares `go 1.25.0`.

## Typical Workflows
### Protobuf Schema Pipeline
1. Author shared `.proto` schemas for RPC, persistence, snapshots, config, or other cross-process data contracts.
2. Run `protoc` and generate the bindings needed by each runtime, such as Go server code, clone helpers, GAP variant integration, or Godot GDScript client code.
3. Keep generated outputs in per-runtime directories, for example `./server/src/gen` and `./client/script/gen`.

### Excel Table Pipeline
1. Author `.xlsx` workbooks.
2. Run `excelc proto` to generate table-oriented `.proto` files.
3. Run the regular Protobuf generation flow on those generated schemas.
4. Run `excelc code` to generate Go or GDScript table access code.
5. Run `excelc data` to export binary and/or JSON table data.

The Excel pipeline is layered on top of Protobuf. Generated table schemas are still ordinary Protobuf messages, so they can also be used for transport, serialization, storage, snapshots, or other tooling beyond table lookup.

### Property Synchronization
1. Define property state types and synchronized methods in Go.
2. Mark declarations for `tools/propc` and generate `*.sync.gen.go`.
3. Declare properties through `addins/propview` so they can load, save, replicate by revision, and synchronize across endpoints.

### Script Hotfixing
1. Install `addins/goscr` on the service side.
2. Point it at one or more local or remote script projects through `goscr.With.Projects(...)`.
3. Build entity prototypes with `goscr.BuildEntityPT(...)` and script metadata.
4. Let the add-in reload script solutions on local file changes or remote source changes.

## Excel End-to-End Example
For a split client/server project, one practical directory layout is:

```text
./config/excel/              # source .xlsx files
./excelc/server/proto/       # server-facing Excel proto schema
./excelc/client/proto/       # client-facing Excel proto schema
./server/src/gen/            # generated server code
./server/res/excel/          # exported server table data
./client/addons/proto/       # copied from tools/protoc-gen-gdscript/godot
./client/addons/excel/       # copied from tools/protoc-gen-gdscript-excel/godot
./client/addons/rpcli/       # copied from godot/rpcli when GAP/RPC is used
./client/script/gen/         # generated client code
./client/excel/              # exported client table data
```

With that layout, the Excel pipeline usually looks like this:

```bash
# 1. Export server-facing proto schema with hash-based unique indexes.
excelc proto \
  --excel_files=./config/excel/Consts.xlsx \
  --excel_dir=./config/excel \
  --pb_out=./excelc/server/proto \
  --pb_package=excel \
  --pb_options=[go_package=./excel] \
  --pb_imports=Consts.proto \
  --pb_unique_index_as=hash_unique_index \
  --targets=s

# 2. Export client-facing proto schema with sorted unique indexes.
excelc proto \
  --excel_files=./config/excel/Consts.xlsx \
  --excel_dir=./config/excel \
  --pb_out=./excelc/client/proto \
  --pb_package=excel \
  --pb_options=[go_package=./excel] \
  --pb_imports=Consts.proto \
  --pb_unique_index_as=sorted_unique_index \
  --targets=c

# 3. Generate server Go protobuf + Excel lookup code.
protoc -I./excelc/server/proto -I./protobuf/include \
  --include_imports \
  --retain_options \
  --go_out=./server/src/gen \
  --go-structure_out=./server/src/gen \
  --go-excel_out=./server/src/gen \
  ./excelc/server/proto/*.proto
excelc code --pb_dir=./excelc/server/proto --pb_package=excel --go_out=./server/src/gen/excel

# 4. Generate client GDScript protobuf + Excel wrapper code.
protoc -I./excelc/client/proto -I./protobuf/include \
  --include_imports \
  --retain_options \
  --gdscript_out=./client/script/gen \
  --gdscript_opt=string_as_string_name=true \
  --gdscript_opt=gap_variant=true \
  --gdscript-excel_out=./client/script/gen \
  ./excelc/client/proto/*.proto
excelc code \
  --pb_dir=./excelc/client/proto \
  --pb_package=excel \
  --gdscript_out=./client/script/gen/excel \
  --gdscript_default_data_dir=res://excel/

# 5. Export server table data.
excelc data \
  --excel_files=./config/excel/Consts.xlsx \
  --excel_dir=./config/excel \
  --pb_dir=./excelc/server/proto \
  --pb_package=excel \
  --targets=s \
  --json_out=./server/res/excel \
  --binary_out=./server/res/excel

# 6. Export client table data as chunked binary files.
excelc data \
  --excel_files=./config/excel/Consts.xlsx \
  --excel_dir=./config/excel \
  --pb_dir=./excelc/client/proto \
  --pb_package=excel \
  --targets=c \
  --binary_out=./client/excel \
  --binary_chunked=true \
  --binary_chunk_size=10000
```

`--targets` does more than choose server/client outputs. It also controls column visibility during Excel export: only fields marked for the selected target are kept in the generated `.proto`, lookup code, and exported table data. In practice, server-only columns can be excluded from client schemas and client data packages, and client-only columns can be omitted from server-side artifacts the same way.

`--pb_unique_index_as` controls how `unique_index` is represented in the exported Protobuf schema. Different values produce different index message shapes, which then affect the generated lookup code and the runtime memory profile of loaded table data. In this example, the server uses `hash_unique_index` so generated Go lookup code can query hash-based indexes, while the client uses `sorted_unique_index` to reduce index memory usage on the Godot side. Generated GDScript wrappers then query `SortedUniqueIndex.Values` with binary search and can still work with `*.bin.idx` / `*.bin.chk_*` chunked table data when needed.

- `hash_unique_index` emits hash-based unique indexes. Generated lookup code can use these indexes for direct key-oriented queries, which is usually a good fit on the server where table data often stays resident in memory.
- `sorted_unique_index` emits unique indexes as sorted arrays. Queries become binary search instead of direct hash lookup, but it usually has lower memory overhead than hash-based indexes, which makes it a better fit for memory-constrained clients.
- `--binary_chunked` controls how binary table data is written during `excelc data`. When enabled, rows are split into `*.bin.chk_*` chunk files and paired with a `*.bin.idx` index file.
- `--binary_chunk_size` controls the maximum row count per chunk. The default value is `10000`. It affects the exported binary layout but does not change the generated Protobuf schema.

## Excel Workbook Specification
### Workbook Structure
- The optional `@types` sheet is used to declare reusable struct and enum types for the workbook. Table sheets can reference those custom types in field definitions instead of being limited to built-in scalar types only.
- In the `@types` sheet, `Meta` supports `separator`, `scope`, and `pb_field_number`. Index-related keys such as `unique_index`, `hash_unique_index`, and `sorted_unique_index` are only meaningful on table columns, not on `@types` struct or enum declarations.
- Any table column whose header name does not start with a letter is ignored by `excelc`. In practice, columns starting with `#` are commonly used as comment columns for notes, examples, or editor-only annotations, and they will not participate in generated schema or exported data.

### Table Sheet Layout
- `tools/excelc/examples/ExampleCN.xlsx` and `tools/excelc/examples/ExampleEN.xlsx` use the same table-sheet structure: row 1 is field names, row 2 is field types, row 3 is field `Meta`, row 4 is comments, and actual data starts from row 5.
- Scalar cells are written directly, such as `1`, `3.14`, `true`, or `HelloWorld`.
- `bytes` cells use Base64 text, as shown by values like `SGVsbG9Xb3JsZA==`.
- Enum cells can use enum number, enum name, or enum alias, such as `1`, `EnumB` / `A`, and localized aliases like `枚举值B`.
- Repeated scalar fields are separated by the field separator. By default this is `,`, so examples look like `1,2,3,4,5`. If `separator=|` is configured in meta, the same field can be written like `HelloWorld|HAHAHAHAHA`.
- Object cells use YAML-style mapping syntax. Example values include `A: 1, B: HelloWorld, C: [1, 2, 3]`, and object fields may also be addressed by alias, such as `FieldA` / `字段A`.
- Repeated object fields support both full YAML sequences like `[{A: 1, B: HelloWorld}, {A: 2, B: HAHAHAHAHA}]` and separator-based forms such as `A: 1, B: HelloWorld | A: 2, B: HAHAHAHAHA` when a custom separator is configured.
- Map cells use YAML-style mapping syntax, for example `0: HelloWorld, 1: HAHAHAHAHA` or `0: {A: 1, B: HelloWorld}, 1: {A: 3, B: HAHAHAHAHA}`.

### Column Parameters
- Field-level options are configured in the table header's `Meta` setting row, also accepted as `元数据` / `特性`. Each field column fills its own meta cell in query-string syntax, such as `scope=c&sorted_unique_index=1` or `separator=|`.
- `scope`: repeatable target visibility tag. It works together with `--targets`; columns without `scope` are visible to all targets. To match multiple targets, repeat the key, for example `scope=c&scope=s` or `scope=client&scope=editor`.
- `separator`: delimiter used when parsing repeated or map-like cell values. The default is `,`.
- `unique_index`: repeatable integer index tag. It defines unique-index groups, and the actual exported representation follows `--pb_unique_index_as`.
- `hash_unique_index`: repeatable integer index tag. It forces the tagged unique index groups to use hash-based representation.
- `sorted_unique_index`: repeatable integer index tag. It forces the tagged unique index groups to use sorted-array representation.
- `pb_field_number`: optional Protobuf field number override. It must be positive, outside Protobuf's reserved field-number range, and unique within the generated message.
- A single-column unique index is configured by assigning one tag on one field, for example `unique_index=1` or `hash_unique_index=1`.
- A composite unique index is configured by reusing the same tag on multiple fields. For example, `role_id` with `hash_unique_index=1` and `level` with `hash_unique_index=1` together form one composite unique index on `(role_id, level)`.
- One table can define multiple unique indexes at the same time by using different tags, for example `id -> hash_unique_index=1`, `name -> sorted_unique_index=2`, and `type + sub_type -> sorted_unique_index=3`.
- One field can participate in multiple indexes by repeating tags in its meta cell, for example `hash_unique_index=1&hash_unique_index=2`. That field will be included in both index groups.
- The same tag cannot appear in both `hash_unique_index` and `sorted_unique_index`.

## Godot Integration
### Runtime Directories
- `tools/protoc-gen-gdscript/godot` is the Protobuf runtime required by every generated `*.pb.gd` file. Copy this directory into the Godot project so global classes such as `ProtoMessage`, `ProtoUtils`, `ProtoInputFile`, and `ProtoOutputBuffer` are available to generated code.
- `tools/protoc-gen-gdscript-excel/godot` is the additional runtime directory used by generated `*.excel.gd` wrappers. Excel wrappers still depend on the Protobuf runtime above, so when using `protoc-gen-gdscript-excel`, both runtime directories must be present in the Godot project.
- `godot/rpcli` is the Godot-side Golaxy RPC client runtime. Copy it into the Godot project when the client connects to Golaxy services through GAP/GTP, or when generated Protobuf code is emitted with `--gdscript_opt=gap_variant=true`.
- `godot/resty` is a lightweight HTTP helper for Godot 4. Copy it into the Godot project when the client needs regular HTTP APIs, file downloads, or SSE streams outside the GAP/GTP RPC channel.

### Layout Rules
- These runtime scripts do not need to live in a fixed directory. A common pattern is to place them under `addons/<name>` and let Godot register them through `class_name`.
- Generated Protobuf scripts that use `gap_variant=true` depend on `GAPVariants` from `godot/rpcli`, so `res://addons/rpcli/` must be installed before those generated files can load.
- Generated `*.pb.gd` files are anonymous scripts by default. Pass `--gdscript_opt=class_name=true` to emit a top-level `class_name` for each generated file script, using names such as `LoginPB`.
- Keep generated `*.pb.gd` files in the same relative layout as the source `.proto` files. Cross-file Protobuf references are emitted as relative `preload(...)` calls.
- Keep application Protobuf output and Excel-derived Protobuf output in different root directories. In practice, communication/storage schemas and table schemas are usually generated and maintained separately.
- Keep each generated `*.excel.gd` file next to its matching `*.pb.gd` file. Generated Excel wrappers preload `./<name>.pb.gd` from the same output directory, and `excelc code --gdscript_out=...` typically emits an aggregate loader such as `tables.gd` into that directory as well.
- The aggregate `tables.gd` script exports `class_name Tables` by default. Use `excelc code --gdscript_class_name=<Name>` to choose a different Godot global class name, or pass an empty value to omit `class_name`.
- `godot/resty` does not depend on the generated Protobuf or Excel runtimes. Register `resty_client.gd` as an autoload when you want a project-wide HTTP client, or instantiate `RestyClient` manually when a scene needs isolated defaults.

### Layout Examples
One common layout for regular Protobuf output:

```text
res://addons/proto/          # files copied from tools/protoc-gen-gdscript/godot
res://script/gen/proto/      # regular protobuf-generated client messages
res://script/gen/proto/login.pb.gd
```

One common layout for Excel table output:

```text
res://addons/proto/          # files copied from tools/protoc-gen-gdscript/godot
res://addons/excel/          # files copied from tools/protoc-gen-gdscript-excel/godot
res://script/gen/excel/      # Excel protobuf + wrapper output
res://script/gen/excel/excelc.pb.gd
res://script/gen/excel/example.pb.gd
res://script/gen/excel/example.excel.gd
res://script/gen/excel/tables.gd
res://excel/                 # exported Excel data files
res://excel/ExampleTable.bin.idx
res://excel/ExampleTable.bin.chk_0
```

One common layout for a Godot client that uses the RPC runtime:

```text
res://addons/rpcli/          # files copied from godot/rpcli
res://addons/proto/          # required when RPC payloads use generated protobuf
res://script/gen/proto/      # optional generated RPC/application messages
```

One common layout for a Godot client that also uses regular HTTP APIs:

```text
res://addons/resty/          # files copied from godot/resty
res://addons/rpcli/          # optional, files copied from godot/rpcli
res://addons/proto/          # optional, required by generated protobuf messages
res://script/gen/proto/      # optional generated RPC/application messages
```

Register the RPC runtime as a Godot autoload:

```ini
[autoload]

RPCli="*res://addons/rpcli/golaxy_rpcli.gd"
```

Register the HTTP runtime as a Godot autoload:

```ini
[autoload]

Resty="*res://addons/resty/resty_client.gd"
```

Then call HTTP APIs from GDScript:

```gdscript
var res := await (
    Resty.set_base_url("https://api.example.com")
    .r()
    .set_bearer_auth("token")
    .set_query_param("page", 1)
    .get_async("/users")
)

if res.is_success():
    print(res.json)
else:
    push_error(res.error_message)
```

Then connect from GDScript:

```gdscript
var ok := await RPCli.connect_to_async(
    "ws://127.0.0.1:8080",
    GolaxyClient.PROTOCOL_WEBSOCKET,
    "user_id",
    "token"
)
```

If the project also uses generated Excel table wrappers, register the aggregate table script as an autoload too:

```ini
[autoload]

Excel="*res://script/gen/excel/tables.gd"
RPCli="*res://addons/rpcli/golaxy_rpcli.gd"
Resty="*res://addons/resty/resty_client.gd"
```

`Resty.r()` creates an independent request snapshot from the current client defaults, including base URL, headers, query parameters, timeout, gzip, redirect, body-size, download-chunk, JSON parsing, and thread settings. Requests support JSON bodies, form bodies, raw bytes, path parameters, output files, `GET` / `POST` / `PUT` / `PATCH` / `DELETE` / `HEAD`, and both `*_async` and `*_start` styles for concurrent work.

`Resty.sse(url)` creates a long-lived Server-Sent Events stream. It sends `GET`, adds `Accept: text/event-stream` and `Cache-Control: no-cache` when missing, emits `opened`, `event_received`, and `closed`, and can be stopped with `close()`.

## Tool Reference
| Command | Key Options | Notes |
| --- | --- | --- |
| `excelc proto` | `--excel_files`, `--excel_dir`, `--pb_out`, `--pb_package`, `--pb_imports`, `--pb_options`, `--pb_unique_index_as`, `--targets` | Generates table `.proto` files and matching `*.protoset` files from Excel workbooks. Prefer `--excel_files` for explicit inputs. |
| `excelc code` | `--pb_dir`, `--pb_package`, `--go_out`, `--gdscript_out`, `--gdscript_class_name`, `--gdscript_default_data_dir` | Generates Go or GDScript table access code from Excel proto files. |
| `excelc data` | `--excel_files`, `--excel_dir`, `--pb_dir`, `--pb_package`, `--targets`, `--binary_out`, `--binary_chunked`, `--binary_chunk_size`, `--json_out`, `--json_multiline`, `--json_indent` | Exports binary or JSON table data from Excel workbooks using the generated proto descriptors. |
| `propc` | `--decl_file` | Reads a property declaration file and writes the sibling `*.sync.gen.go`. The default comes from `GOFILE`, which makes it convenient for `go generate`. |
| `protoc-gen-gdscript` | `string_as_string_name`, `gap_variant`, `class_name` | Passed as `--gdscript_opt=<name>=<value>` to control string mapping, GAP variant integration, and Godot `class_name` export. |
| `protoc-gen-gdscript-excel` | `string_as_string_name` | Passed as `--gdscript-excel_opt=<name>=<value>` to control string field mapping in Excel wrappers. |

## Package Layout
| Path | Responsibility |
| --- | --- |
| [`./addins/goscr`](./addins/goscr) | Service-level Go scripting add-in, script-backed entity/component helpers, lifecycle bridges. |
| [`./addins/goscr/dynamic`](./addins/goscr/dynamic) | Dynamic script loading, project/solution management, and hotfix support. |
| [`./addins/goscr/fwlib`](./addins/goscr/fwlib) | Symbols exported into script runtimes for `core`, `framework`, and `scaffold` packages. |
| [`./addins/propview`](./addins/propview) | Runtime property view, property sync, marshaling, and replication helpers. |
| [`./tools/excelc`](./tools/excelc) | Excel schema generation, access-code generation, and data export CLI. |
| [`./tools/excelc/examples`](./tools/excelc/examples) | Example workbooks for the Excel table pipeline. |
| [`./tools/excelc/excelutils`](./tools/excelc/excelutils) | Hash/index conversion, table loading, and lookup helpers used by generated code. |
| [`./tools/propc`](./tools/propc) | Property synchronization code generator. |
| [`./tools/protoc-gen-go-excel`](./tools/protoc-gen-go-excel) | Go Protobuf plugin for Excel table lookup APIs. |
| [`./tools/protoc-gen-go-structure`](./tools/protoc-gen-go-structure) | Go Protobuf plugin for clone helpers. |
| [`./tools/protoc-gen-go-variant`](./tools/protoc-gen-go-variant) | Go Protobuf plugin for GAP variant integration. |
| [`./tools/protoc-gen-gdscript`](./tools/protoc-gen-gdscript) | GDScript Protobuf plugin for message types and serialization. |
| [`./tools/protoc-gen-gdscript/godot`](./tools/protoc-gen-gdscript/godot) | Godot Protobuf runtime scripts required by generated `*.pb.gd` files. |
| [`./tools/protoc-gen-gdscript-excel`](./tools/protoc-gen-gdscript-excel) | GDScript Protobuf plugin for Excel table wrappers. |
| [`./tools/protoc-gen-gdscript-excel/godot`](./tools/protoc-gen-gdscript-excel/godot) | Godot Excel runtime scripts required by generated `*.excel.gd` wrappers. |
| [`./godot/rpcli`](./godot/rpcli) | Godot Golaxy RPC client runtime for GAP/GTP connections and callbacks. |
| [`./godot/resty`](./godot/resty) | Godot Resty-style HTTP client runtime for HTTP requests, downloads, concurrent handles, and SSE streams. |

## Related Repositories
- [Golaxy Distributed Service Development Framework Core](https://github.com/pangdogs/core)
- [Golaxy Distributed Service Development Framework](https://github.com/pangdogs/framework)
