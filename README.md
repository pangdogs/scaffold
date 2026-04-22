# Scaffold
[English](./README.md) | [简体中文](./README.zh_CN.md)

## Overview
`scaffold` is the tooling and add-in companion to
[**Golaxy Distributed Service Development Framework**](https://github.com/pangdogs/framework).
It focuses on the project-scaffolding parts that sit around the main runtime:
Excel table compilation, protobuf/GDScript code generation, Go-script hotfix
support, and entity property synchronization.

The repository is organized around two layers:

- `addins`: service/runtime extensions that plug into `git.golaxy.org/framework`.
- `tools`: command-line generators and protobuf plugins used during build-time
  code and data generation.

## What This Module Provides
- `addins/goscr`: a service add-in built on Yaegi for loading Go scripts,
  declaring script-backed entities/components, and hot-reloading local or
  remote script projects.
- `addins/propview`: a runtime add-in for managed entity properties, including
  load/save/sync flows across services and gateway-facing clients.
- `tools/propc`: a property-sync generator that scans annotated Go declarations
  and emits `propview` wrappers for synchronized operations.
- `tools/excelc`: an Excel pipeline that turns `.xlsx` tables into proto
  schemas, generated access code, and exported binary/JSON data files.
- `tools/protoc-gen-go-excel`: a protobuf plugin that adds table lookup and
  index-based access helpers to generated Go code.
- `tools/protoc-gen-go-structure`: a protobuf plugin that adds deep-clone
  helpers for messages, slices, maps, bytes, and nested messages.
- `tools/protoc-gen-go-variant`: a protobuf plugin that makes generated
  messages usable as GAP variant values in the Golaxy RPC stack.
- `tools/protoc-gen-gdscript`: a protobuf plugin that emits GDScript message
  types and serialization logic for Godot-side integration.
- `tools/protoc-gen-gdscript-excel`: a protobuf plugin that emits GDScript
  table wrappers and lookup helpers for Excel-generated schemas.

## Typical Workflows
### Protobuf schema pipeline
1. Author shared `.proto` schemas for RPC, persistence, snapshots, config, or
   any other cross-process data contract.
2. Run `protoc` and generate the bindings you need for each runtime, such as
   Go server code, clone helpers, GAP variant integration, or GDScript client
   code.
3. Keep generated outputs in per-runtime directories, for example
   `./server/src/gen` for Go and `./client/script/gen` for Godot.

### Excel table pipeline
1. Author `.xlsx` tables.
2. Run `excelc proto` to generate table-oriented `.proto` files.
3. Run the normal protobuf pipeline on those generated schemas.
4. Run `excelc data` to export binary and/or JSON table data.

The Excel pipeline is layered on top of protobuf. The generated table schemas
are still ordinary protobuf messages, so they can also be used for transport,
serialization, storage, snapshots, or other tooling beyond table lookup.

### Property synchronization
1. Define property state types and synchronized methods in Go.
2. Mark declarations for `tools/propc` and generate `*.sync.gen.go`.
3. Declare properties through `addins/propview` so they can load, save, and
   replicate revisions across services or clients.

### Script hotfixing
1. Install `addins/goscr` on the service side.
2. Point it at one or more local or remote script projects through
   `goscr.With.Projects(...)`.
3. Build entity prototypes with `goscr.BuildEntityPT(...)` and script metadata.
4. Let the add-in reload script solutions on local file changes or remote
   source changes.

## Godot Runtime Libraries
### Protobuf Runtime
- `tools/protoc-gen-gdscript/libs` is the runtime required by every generated
  `*.pb.gd` file. Copy this directory into your Godot project so global
  classes such as `ProtoMessage`, `ProtoUtils`, `ProtoInputFile`, and
  `ProtoOutputBuffer` are available to generated protobuf code.

### Excel Table Runtime
- `tools/protoc-gen-gdscript-excel/libs` is the additional runtime directory
  used by generated `*.excel.gd` wrappers.
- Excel wrappers still depend on the protobuf runtime above, so when using
  `protoc-gen-gdscript-excel`, both runtime directories must be present in the
  Godot project.

### Layout Rules
- These runtime scripts do not need to live in a fixed directory. A common
  pattern is to place them under `libs` or `addons/<name>` and let Godot
  register them through `class_name`.
- Keep generated `*.pb.gd` files in the same relative layout as the source
  `.proto` files. Cross-file protobuf references are emitted as relative
  `preload(...)` calls.
- Keep application protobuf output and Excel-derived protobuf output in
  different root directories. In practice, communication/storage schemas and
  table schemas are usually generated and maintained separately.
- Keep each generated `*.excel.gd` file next to its matching `*.pb.gd` file.
  Generated Excel wrappers preload `./<name>.pb.gd` from the same output
  directory, and `excelc code --gdscript_out=...` typically emits an aggregate
  loader such as `tables.gd` into that directory as well.

## Example Godot Layout
### Protobuf Pipeline
One common layout for regular protobuf output inside a Godot project:

```text
res://addons/proto/          # files copied from tools/protoc-gen-gdscript/libs
res://script/gen/proto/      # regular protobuf-generated client messages
res://script/gen/proto/login.pb.gd
```

### Excel Table Pipeline
One common layout for Excel table output inside a Godot project.
This section only lists the Excel table side of the layout:

```text
res://addons/excel/          # files copied from tools/protoc-gen-gdscript-excel/libs
res://script/gen/excel/      # excel protobuf + wrapper output
res://script/gen/excel/excelc.pb.gd
res://script/gen/excel/example.pb.gd
res://script/gen/excel/example.excel.gd
res://script/gen/excel/tables.gd
res://res/excel/             # exported excel data files
res://res/excel/ExampleTable.bin.idx
res://res/excel/ExampleTable.bin.chk_0
```

For a split client/server project, one practical directory layout is:

```text
./config/excel/              # source .xlsx files
./excelc/server/proto/       # server-facing excel proto schema
./excelc/client/proto/       # client-facing excel proto schema
./server/src/gen/            # generated Go protobuf/plugin output
./server/res/excel/          # exported server table data
./client/script/gen/         # generated GDScript protobuf/plugin output
./client/res/excel/          # exported client table data
```

With a layout like that, the Excel pipeline usually looks like this:

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
  --gdscript-excel_out=./client/script/gen \
  ./excelc/client/proto/*.proto
excelc code --pb_dir=./excelc/client/proto --pb_package=excel --gdscript_out=./client/script/gen/excel

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
  --binary_out=./client/res/excel \
  --binary_chunked=true
```

`--targets` does more than choose server/client outputs. It also controls
column visibility during Excel export: only fields marked for the selected
target are kept in the generated `.proto`, lookup code, and exported table
data. In practice, server-only columns can be excluded from client schemas and
client data packages, and client-only columns can be omitted from server-side
artifacts the same way.

`--pb_unique_index_as` controls how unique indexes are represented in the
exported protobuf schema. Different values produce different index message
shapes, which then affects the generated lookup code and the runtime memory
profile of loaded table data. In this example, the split is intentional: the
server keeps `hash_unique_index` so generated Go lookup code can query
hash-based indexes, while the client uses `sorted_unique_index` to reduce
index memory usage on the Godot side. Generated GDScript wrappers then query
`SortedUniqueIndex.Values` with binary search and can still work with
`*.bin.idx` / `*.bin.chk_*` chunked table data when needed.

- `hash_unique_index` emits hash-based unique indexes. Generated lookup code
  can use these indexes for direct key-oriented queries, which is usually a
  good fit on the server where table data often stays resident in memory.
- `sorted_unique_index` emits unique indexes as sorted arrays. Queries become
  binary search instead of direct hash lookup, but it usually has lower memory
  overhead than hash-based indexes, which makes it a better fit for
  memory-constrained clients.
- In practice, `hash_unique_index` is a good default for server-side lookup
  code, while `sorted_unique_index` is a better fit for Godot clients that
  need to keep table index memory usage under control.

`--binary_chunked` controls how binary table data is written during
`excelc data`. When enabled, rows are split into `*.bin.chk_*` chunk files and
paired with a `*.bin.idx` index file, which helps large client tables avoid
loading one monolithic binary blob at once. It affects the exported binary
layout, but does not change the generated protobuf schema.

Excel workbook conventions:

- The optional `@types` sheet is used to declare reusable struct and enum
  types for the workbook. Table sheets can then reference those custom types in
  field definitions, instead of being limited to built-in scalar types only.
- In the `@types` sheet, `Meta` effectively supports `separator` and `scope`.
  Index-related keys such as `unique_index`, `hash_unique_index`, and
  `sorted_unique_index` are only meaningful on table columns, not on `@types`
  struct or enum declarations.
- Any table column whose header name does not start with a letter is ignored by
  `excelc`. In practice, columns starting with `#` are commonly used as comment
  columns for notes, examples, or editor-only annotations, and they will not
  participate in generated schema or exported data.

Table sheet data layout:

- `ExampleCN.xlsx` and `ExampleEN.xlsx` use the same table-sheet structure:
  row 1 is field names, row 2 is field types, row 3 is field meta, row 4 is
  comments, and actual data starts from row 5.
- Scalar cells are written directly, such as `1`, `3.14`, `true`, or
  `HelloWorld`.
- `bytes` cells use Base64 text, as shown by values like
  `SGVsbG9Xb3JsZA==`.
- Enum cells can use enum number, enum name, or enum alias. The examples show
  all three forms such as `1`, `EnumB` / `A`, and localized aliases like
  `枚举值B`.
- Repeated scalar fields are separated by the field separator. By default this
  is `,`, so examples look like `1,2,3,4,5`. If `separator=|` is configured in
  meta, the same field can be written like `HelloWorld|HAHAHAHAHA`.
- Object cells use YAML-style mapping syntax. Example values include
  `A: 1, B: HelloWorld, C: [1, 2, 3]`, and object fields may also be addressed
  by alias, such as `FieldA` / `字段A`.
- Repeated object fields support both full YAML sequences like
  `[{A: 1, B: HelloWorld}, {A: 2, B: HAHAHAHAHA}]` and separator-based forms
  such as `A: 1, B: HelloWorld | A: 2, B: HAHAHAHAHA` when a custom separator
  is configured.
- Map cells use YAML-style mapping syntax, for example
  `0: HelloWorld, 1: HAHAHAHAHA` or
  `0: {A: 1, B: HelloWorld}, 1: {A: 3, B: HAHAHAHAHA}`.

Excel column parameters:

- Field-level options are configured in the table header's `Meta` setting row
  (also accepted as `元数据` / `特性`). Each field column fills its own meta
  cell in that header row, using query-string syntax such as
  `scope=c&sorted_unique_index=1` or `separator=|`.
- `scope`: repeatable target visibility tag. It works together with
  `--targets`; columns without `scope` are visible to all targets. To match
  multiple targets, repeat the key, for example
  `scope=c&scope=s` or `scope=client&scope=editor`.
- `separator`: delimiter used when parsing repeated or map-like cell values.
  The default is `,`.
- `unique_index`: repeatable integer index tag. It defines unique-index groups,
  and the actual exported representation follows `--pb_unique_index_as`.
- `hash_unique_index`: repeatable integer index tag. It forces the tagged
  unique index groups to use hash-based representation.
- `sorted_unique_index`: repeatable integer index tag. It forces the tagged
  unique index groups to use sorted-array representation.
- A single-column unique index is configured by assigning one tag on one field,
  for example `unique_index=1` or `hash_unique_index=1`.
- A composite unique index is configured by reusing the same tag on multiple
  fields. For example, `role_id` with `hash_unique_index=1` and `level` with
  `hash_unique_index=1` together form one composite unique index on
  `(role_id, level)`.
- One table can define multiple unique indexes at the same time by using
  different tags, for example `id -> hash_unique_index=1`,
  `name -> sorted_unique_index=2`, and
  `type + sub_type -> sorted_unique_index=3`.
- One field can participate in multiple indexes by repeating tags in its meta
  cell, for example `hash_unique_index=1&hash_unique_index=2`. That field will
  be included in both index groups.
- The same tag cannot appear in both `hash_unique_index` and
  `sorted_unique_index`.

## Package Layout
| Path | Responsibility |
| --- | --- |
| [`./addins/goscr`](./addins/goscr) | Service-level Go scripting add-in, script-backed entity/component helpers, lifecycle bridges |
| [`./addins/goscr/dynamic`](./addins/goscr/dynamic) | Dynamic script loading, project/solution management, and hotfix support |
| [`./addins/goscr/fwlib`](./addins/goscr/fwlib) | Symbols exported into script runtimes for `core`, `framework`, and `scaffold` packages |
| [`./addins/propview`](./addins/propview) | Runtime property view, property sync, marshaling, and replication helpers |
| [`./tools/excelc`](./tools/excelc) | Excel schema generation, access-code generation, and data export CLI |
| [`./tools/excelc/excelutils`](./tools/excelc/excelutils) | Hash/index conversion, table loading, and lookup helpers used by generated code |
| [`./tools/propc`](./tools/propc) | Property synchronization code generator |
| [`./tools/protoc-gen-go-excel`](./tools/protoc-gen-go-excel) | Go protobuf plugin for Excel table lookup APIs |
| [`./tools/protoc-gen-go-structure`](./tools/protoc-gen-go-structure) | Go protobuf plugin for clone helpers |
| [`./tools/protoc-gen-go-variant`](./tools/protoc-gen-go-variant) | Go protobuf plugin for GAP variant integration |
| [`./tools/protoc-gen-gdscript`](./tools/protoc-gen-gdscript) | GDScript protobuf plugin for message types and serialization |
| [`./tools/protoc-gen-gdscript/libs`](./tools/protoc-gen-gdscript/libs) | Godot protobuf runtime scripts required by generated `*.pb.gd` files |
| [`./tools/protoc-gen-gdscript-excel`](./tools/protoc-gen-gdscript-excel) | GDScript protobuf plugin for Excel table wrappers |
| [`./tools/protoc-gen-gdscript-excel/libs`](./tools/protoc-gen-gdscript-excel/libs) | Godot Excel runtime scripts required by generated `*.excel.gd` wrappers |

## Toolchain Notes
- `tools/excelc` uses Cobra/Viper and is split into three subcommands:
  `proto`, `code`, and `data`.
- `tools/propc` reads a Go declaration file and writes a sibling
  `*.sync.gen.go` file.
- `tools/excelc/examples` contains example workbooks for the Excel pipeline.

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

## Related Repositories
- [Golaxy Distributed Service Development Framework Core](https://github.com/pangdogs/core)
- [Golaxy Distributed Service Development Framework](https://github.com/pangdogs/framework)
