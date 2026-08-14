# Scaffold

[English](./README.md) | [简体中文](./README.zh_CN.md)

## Overview

`scaffold` is the companion tools and add-ins repository for the [Golaxy Distributed Service Development Framework](https://github.com/pangdogs/framework). It provides code generation, data export, runtime integration, property synchronization, and script hot-reloading for Go servers, Godot clients, and Excel-based configuration pipelines.

This repository is not a standalone application framework. It contains three kinds of components:

- `tools`: build-time CLIs and `protoc` plugins.
- `addins`: Go runtime extensions for `git.golaxy.org/framework`.
- `godot`: client runtime scripts that can be copied into a Godot 4 project.

## Documentation Map

- [Components](#components)
- [Requirements and Installation](#requirements-and-installation)
- [Protobuf Code Generation](#protobuf-code-generation)
- [Excel Table Pipeline](#excel-table-pipeline)
- [Property Synchronization Generation](#property-synchronization-generation)
- [Runtime Components](#runtime-components)
- [Repository Layout](#repository-layout)

## Components

| Kind          | Component                               | Responsibility                                                                                                                   |
|---------------|-----------------------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| Go add-in     | `addins/goscr`                          | Loads Yaegi-based Go script projects and supports scripted entities/components, local or remote source updates, and hot reloads. |
| Go add-in     | `addins/propview`                       | Manages entity property loading, persistence, revisions, and replication across services or clients.                             |
| CLI           | `tools/propc`                           | Scans annotated Go property declarations and generates `*.sync.gen.go`.                                                          |
| CLI           | `tools/excelc`                          | Generates table proto schemas, aggregate access code, and JSON/binary data from `.xlsx` files.                                   |
| protoc plugin | `tools/protoc-gen-go-structure`         | Generates deep-copy helpers for Go Protobuf messages.                                                                            |
| protoc plugin | `tools/protoc-gen-go-variant`           | Makes Go Protobuf messages implement the Golaxy GAP variant contract.                                                            |
| protoc plugin | `tools/protoc-gen-go-excel`             | Generates Go table lookup and index methods for Excel schemas.                                                                   |
| protoc plugin | `tools/protoc-gen-gdscript`             | Generates Godot-facing `*.pb.gd` Protobuf messages.                                                                              |
| protoc plugin | `tools/protoc-gen-gdscript-excel`       | Generates Godot-facing `*.excel.gd` table wrappers and index queries.                                                            |
| Godot runtime | `tools/protoc-gen-gdscript/godot`       | Protobuf codecs, streams, hashing, and base message classes used by `*.pb.gd`.                                                   |
| Godot runtime | `tools/protoc-gen-gdscript-excel/godot` | Index lookup, chunk-file, and binary-search helpers used by `*.excel.gd`.                                                        |
| Godot runtime | `godot/rpcli`                           | GAP/GTP connections, reconnects, RPC calls, callbacks, and variant transport.                                                    |
| Godot runtime | `godot/resty`                           | Resty-style HTTP requests, downloads, concurrent requests, and Server-Sent Events.                                               |

## Requirements and Installation

### Prerequisites

- Go `1.25.0`, or a version compatible with the current `go.mod`.
- `protoc` and an include directory containing `google/protobuf/descriptor.proto`.
- The official `protoc-gen-go` when generating Go Protobuf bindings.
- Godot 4 when using generated GDScript and its runtime scripts.

### Install `protoc`

`protoc` is the Protocol Buffers compiler distributed by the official [protocolbuffers/protobuf](https://github.com/protocolbuffers/protobuf) project. Use it to compile the `.proto` schemas emitted by `excelc proto`, generate Protobuf bindings, and create the `*.protoset` descriptor sets used by the Excel workflow. In particular, the generated `excelc.proto` imports `google/protobuf/descriptor.proto`, so retain the compiler distribution's `include` directory.

On Windows, install it with Winget and open a new terminal before verifying the installation:

```powershell
winget install protobuf
protoc --version
```

Alternatively, download the precompiled archive for the target operating system and architecture from the [official releases](https://github.com/protocolbuffers/protobuf/releases/latest), for example `protoc-<version>-win64.zip`. Extract it to a custom root (the example uses `C:\tools\protobuf`), keep both its `bin` and `include` directories, and add `bin` to `PATH`:

```text
C:\tools\protobuf\
├─ bin\
│  └─ protoc.exe
└─ include\
   └─ google\protobuf\descriptor.proto
```

The following PowerShell commands configure and validate a manually extracted Windows installation for the current terminal:

```powershell
$protocRoot = 'C:\tools\protobuf'
$protocBin = Join-Path $protocRoot 'bin'
$env:Path = "$protocBin;$env:Path"
$env:PROTOBUF_INCLUDE = Join-Path $protocRoot 'include'

protoc --version
Test-Path (Join-Path $env:PROTOBUF_INCLUDE 'google\protobuf\descriptor.proto')
```

Set `PROTOBUF_INCLUDE` (or pass the same directory with `-I`) to the `include` root, not to its `google/protobuf` subdirectory. On macOS and Debian/Ubuntu, the corresponding package-manager commands are `brew install protobuf` and `sudo apt-get install protobuf-compiler`; verify the resulting compiler with `protoc --version`.

### Add the Go module

Add this repository as a Go module dependency when application code imports the add-ins or generated runtime helpers:

```bash
go get git.golaxy.org/scaffold@latest
```

### Install code generators

Code generators are normally installed with Go. If users should not need to install Go after a submission or release, the compiled tools can instead be archived in the project directory.

#### Install with Go

For development machines, run the `go install` commands below. When `GOBIN` is unset, Go usually installs the compiled executables to `$(go env GOPATH)/bin`; make sure that directory and the `bin` directory containing `protoc` are both on `PATH`.

To install the Go tools into a custom directory, set `GOBIN` before running the commands, for example:

```powershell
$generatorBin = 'C:\tools\scaffold\bin'
New-Item -ItemType Directory -Force -Path $generatorBin | Out-Null
$env:GOBIN = $generatorBin
$env:Path = "$generatorBin;$env:Path"
```

Then run:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install git.golaxy.org/scaffold/tools/excelc@latest
go install git.golaxy.org/scaffold/tools/propc@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-excel@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-structure@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-variant@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-gdscript@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-gdscript-excel@latest
```

#### Archive prebuilt tools

If users should not need to install Go after a submission or release, copy the already compiled `.exe` files into the project's `tools` directory. A layout matching `garden/tools` keeps `protoc.exe` and `protoc-gen-*` plugins in `protobuf/bin`, `excelc.exe` in `excelc/bin`, and (when used) `propc.exe` in a separate `propc/bin`. Get `protoc.exe` and `include` from the official `protocolbuffers/protobuf` release archive.

```text
tools\
├─ protobuf\
│  ├─ bin\
│  │  ├─ protoc.exe
│  │  ├─ protoc-gen-go.exe
│  │  ├─ protoc-gen-go-excel.exe
│  │  ├─ protoc-gen-go-structure.exe
│  │  ├─ protoc-gen-go-variant.exe
│  │  ├─ protoc-gen-gdscript.exe
│  │  └─ protoc-gen-gdscript-excel.exe
│  └─ include\
│     └─ google\protobuf\descriptor.proto
├─ excelc\
│  └─ bin\excelc.exe
└─ propc\
   └─ bin\propc.exe
```

From the project root, run the following PowerShell commands to use these project-local tools in the current terminal:

```powershell
$projectTools = Join-Path (Get-Location) 'tools'
$protobufBin = Join-Path $projectTools 'protobuf\bin'
$excelcBin = Join-Path $projectTools 'excelc\bin'
$propcBin = Join-Path $projectTools 'propc\bin'
$env:Path = "$protobufBin;$excelcBin;$propcBin;$env:Path"
$env:PROTOBUF_INCLUDE = Join-Path $projectTools 'protobuf\include'
```

Make sure the project-local directories containing `protoc`, `excelc`, and `propc` are on `PATH`. `protoc` maps `protoc-gen-<name>` to `--<name>_out`; pass plugin options with `--<name>_opt`.

The Protobuf plugins in this repository use Go `protogen`. Even for GDScript-only output, every input `.proto` needs a valid `option go_package`, or an `M<file>=<go-import-path>` plugin parameter that supplies its Go import path.

## Protobuf Code Generation

### Minimal Schema

This example demonstrates regular application Protobuf generation. See the [complete Excel workflow](#complete-workflow) for schemas emitted by `excelc`.

```proto
syntax = "proto3";

package example;
option go_package = "example.com/project/server/gen/pb;pb";

message Profile {
  int64 id = 1;
  string name = 2;
  repeated int32 tags = 3;
}
```

Assume it is stored at `proto/example/profile.proto`.

### Generate Go Code

```bash
protoc \
  -I./proto \
  --go_out=./server/gen/pb \
  --go_opt=paths=source_relative \
  --go-structure_out=./server/gen/pb \
  --go-structure_opt=paths=source_relative \
  --go-variant_out=./server/gen/pb \
  --go-variant_opt=paths=source_relative \
  --go-variant_opt=deterministic=true \
  ./proto/example/profile.proto
```

The command produces:

| File                   | Generator                 | Contents                                            |
|------------------------|---------------------------|-----------------------------------------------------|
| `profile.pb.go`        | `protoc-gen-go`           | Official Go Protobuf message.                       |
| `profile.structure.go` | `protoc-gen-go-structure` | Message- and field-level deep-copy helpers.         |
| `profile.variant.go`   | `protoc-gen-go-variant`   | GAP variant registration, type id, and I/O methods. |

### `protoc-gen-go-structure`

Run this plugin against the same schemas and Go package as `protoc-gen-go`. It does not define messages; for every top-level message in a file, it adds:

- A `Clone()` deep copy.
- Field-level `Clone<Field>()` helpers.
- Appropriate cloning for messages, lists, maps, `bytes`, and scalar fields.

The plugin has no custom options. It accepts standard `protogen` parameters such as `paths`, `module`, and `M<file>=<import>`.

### `protoc-gen-go-variant`

This plugin makes every top-level Go message in a file participate in the Golaxy GAP variant system and emits `*.variant.go` with:

- Message registration in `init()`.
- `Read`, `Write`, `Size`, `TypeId`, and `Indirect` methods.
- A stable custom variant type id derived from the Protobuf package and message name.

| Option          | Default | Description                                                                                                     |
|-----------------|---------|-----------------------------------------------------------------------------------------------------------------|
| `deterministic` | `false` | Uses deterministic Protobuf serialization. Enable it when stable bytes or hashes are required across processes. |

Generated code depends on `git.golaxy.org/framework/net/gap/variant`.

Go and GDScript variants use compatible 32-bit FNV-1a type ids. The id is stable for the same `package.message`, but the generators do not detect hash collisions across the complete schema set. Large protocols should validate type-id uniqueness during builds or tests.

### Generate GDScript Code

```bash
protoc \
  -I./proto \
  --gdscript_out=./client/script/gen \
  --gdscript_opt=paths=source_relative \
  --gdscript_opt=string_as_string_name=true \
  --gdscript_opt=deterministic=true \
  ./proto/example/profile.proto
```

This emits `profile.pb.gd`. Generated scripts are not self-contained: copy every file from [`tools/protoc-gen-gdscript/godot`](./tools/protoc-gen-gdscript/godot) into the Godot project, for example under `res://addons/proto/`.

### `protoc-gen-gdscript`

For proto3 enums and messages, this plugin generates:

- GDScript fields, nested messages, and enums.
- Binary serialization, deserialization, and size calculation.
- `reset`, `clone`, `equals`, `hash`, `to_dict`, and `from_dict`.
- Repeated fields, maps, packed/unpacked numeric fields, and cross-file preloads.

| Option                  | Default | Description                                                                                                 |
|-------------------------|---------|-------------------------------------------------------------------------------------------------------------|
| `string_as_string_name` | `false` | Maps proto `string` to `StringName`. This is useful for repeated identifiers, not arbitrary long text.      |
| `deterministic`         | `false` | Sorts map keys before serialization so equivalent messages produce stable bytes.                            |
| `gap_variant`           | `false` | Makes messages extend `ProtoGAPVariant` and registers them with `GAPVariants`; also requires `godot/rpcli`. |
| `class_name`            | `false` | Exports a top-level `class_name` from generated file scripts; scripts use `preload` by default.             |

Cross-file references use relative `preload(...)`, so the output should preserve the relative source `.proto` layout. When `gap_variant=true` is enabled, also install [`godot/rpcli`](./godot/rpcli) so `GAPVariants` is available.

#### Protobuf Support Scope

`protoc-gen-gdscript` only targets `syntax = "proto3"`; proto2 and Protobuf Editions are unsupported. Ordinary scalars, enums, messages, `repeated`, `map`, and packed/unpacked arrays work, with these boundaries:

| Feature               | Current behavior and limitations                                                                                                                                                                                          |
|-----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Unknown fields        | Unknown varint, fixed32, fixed64, and length-delimited fields can be skipped but are not retained, so they disappear after reserialization. Unknown group fields cannot be skipped.                                       |
| `oneof`               | No case/discriminator, mutual exclusion, or automatic clearing is generated. Members are treated as independent fields and should not be used for GDScript targets.                                                       |
| `optional` / presence | Proto3 scalar `optional` fields have no `has_*` / `clear_*` state; absent and explicitly assigned default values cannot be distinguished. Message absence can still be represented by `null`.                             |
| Services              | No RPC client/server stubs are generated from `service` declarations.                                                                                                                                                     |
| ProtoJSON             | `to_dict` / `from_dict` are convenience conversions, not a complete ProtoJSON implementation. Special mappings for `Any`, `Timestamp`, `Duration`, `FieldMask`, `Struct`, and other well-known types are not implemented. |
| `uint64` / `fixed64`  | Wire encoding preserves all 64 bits, but Godot `int` is signed 64-bit. Values above `9223372036854775807` appear negative; keep business values in the positive int64 range when practical.                               |

## Excel Table Pipeline

### Tool and Artifact Flow

The Excel toolchain has schema, code, and data stages. `excelc proto` derives Protobuf schemas from workbooks, `protoc` and its plugins generate per-table language bindings, `excelc code` generates the aggregate entry points, and `excelc data` produces the runtime data:

```text
.xlsx
  │
  ├─ excelc proto ───────────────> excelc.proto + <Workbook>.proto
  │                                  │
  │                                  └─ protoc + plugins
  │                                      ├─ *.pb.go / *.structure.go / *.excel.go
  │                                      ├─ *.pb.gd / *.excel.gd
  │                                      └─ *.protoset
  │
  ├─ excelc code + *.protoset ───> tables.go / tables.gd
  │
  └─ excelc data + *.protoset ───> *.json / *.bin / *.bin.idx + *.bin.chk_*
```

The easily missed artifact is `*.protoset`: `excelc proto` only emits `.proto`. Afterwards, run `protoc --descriptor_set_out` separately for `excelc.proto` and every workbook proto, creating a matching `.protoset`. Both `excelc code` and `excelc data` reconstruct dynamic messages and custom options from those descriptor sets.

For example, `Config.xlsx` produces `Config.proto`, which declares the row message `ConfigColumns` and the table message `ConfigTable` by default. Its per-table code files are `Config.pb.go`, `Config.structure.go`, and `Config.excel.go` (or `Config.pb.gd` and `Config.excel.gd` for Godot). Runtime data uses the table message name, producing `ConfigTable.json`, `ConfigTable.bin`, or `ConfigTable.bin.idx`.

#### Artifact Responsibilities

| Artifact | Producer | Responsibility |
|----------|----------|----------------|
| `excelc.proto` | `excelc proto` | Shared declarations for every table, including Excel custom options, index structures, and chunk manifests. It is emitted once per target schema set. |
| `<Workbook>.proto` | `excelc proto` | Static schema for one workbook: `*Columns`, `*Table`, workbook-local messages and enums, and table index fields. It contains no data rows. |
| `*.protoset` | `protoc --descriptor_set_out` | Descriptor sets used by `excelc code` and `excelc data` to recover messages, fields, scopes, and index options. They are build-time inputs only. |
| `*.pb.go` | `protoc-gen-go` | Go Protobuf messages, enums, and wire-format types. Exported data is deserialized into the generated `*Table` and `*Columns` types. |
| `*.structure.go` | `protoc-gen-go-structure` | Optional Go deep-copy and field-cloning helpers; it does not load or query tables. |
| `*.excel.go` | `protoc-gen-go-excel` | Per-table Go lookup code, adding `Lookup`, `Get`, and `LookupBy...` methods for unique and non-unique indexes to `*Table`. |
| `tables.go` | `excelc code --go_out` | Aggregate Go entry point. `Tables` has one field per table; `LoadJsonFiles` and `LoadBinaryFiles` load every table from one directory and return that container. |
| `*.pb.gd` | `protoc-gen-gdscript` | Godot Protobuf messages, enums, serialization, and deserialization, corresponding to `*.pb.go`. |
| `*.excel.gd` | `protoc-gen-gdscript-excel` | Per-table GDScript wrapper with row access, index queries, and synchronous/asynchronous chunked-table access. |
| `tables.gd` | `excelc code --gdscript_out` | Aggregate Godot entry point. It preloads each `*.pb.gd` / `*.excel.gd`, re-exports messages and enums, owns one wrapper per table, and loads ordinary or chunked binaries through `load_data()`. It is commonly registered as an autoload. |
| `*Table.json` | `excelc data --json_out` | Readable Protobuf JSON containing rows and indexes, primarily for Go `LoadJsonFiles`, inspection, and hot loading. |
| `*Table.bin` | `excelc data --binary_out` | Complete binary Protobuf table message containing both rows and indexes; Go and Godot can load it. |
| `*Table.bin.idx` | `excelc data --binary_out --binary_chunked` | Chunked entry file containing indexes and the chunk manifest, but no `Rows`. |
| `*Table.bin.chk_N` | same command | Chunk data containing only its range of `Rows`; the Godot wrapper loads it on demand for queries or row access. |

Neither `tables.go` nor `tables.gd` embeds table data. They centralize loading all tables and accessing each table. Using a generated `ConfigTable` as an example, the runtime flow is:

```text
Go:
ConfigTable.json / ConfigTable.bin
  └─ LoadJsonFiles / LoadBinaryFiles
       └─ tables *Tables
            └─ tables.ConfigTable.Lookup(...)

Godot, ordinary binary:
Excel.load_data()
  └─ read ConfigTable.bin (rows and indexes)
       └─ create Excel.ConfigTable
            └─ Excel.ConfigTable.lookup(...)

Godot, chunked binary:
Excel.load_data()
  └─ read only ConfigTable.bin.idx (indexes and chunk manifest)
       └─ create Excel.ConfigTable (internally a ConfigChunkedTable)
            └─ Excel.ConfigTable.lookup(...) / await Excel.ConfigTable.lookup_async(...)
                 └─ resolve the row offset from the index
                      └─ load and cache the corresponding ConfigTable.bin.chk_N on demand
```

Here `Excel` is the instance name when `tables.gd` is registered as an autoload; applications that do not use an autoload can create their own `Tables` instance. Both synchronous and asynchronous chunked lookups load chunks on demand. When threads are available, a synchronous call waits for the background load while an asynchronous call yields on the main thread; without thread support, Godot performs the load on the calling thread, so both APIs remain functional but the first access blocks synchronously. Calling `rows()` or `rows_async()` requires all chunks to be loaded.

The `.xlsx`, `.proto`, and `*.protoset` files are configuration or build inputs and normally are not shipped with the application. A Go application compiles the generated `*.go` files and deploys JSON or binary data according to its loader. A Godot application needs the generated `*.gd` files, both Godot runtime script sets, and binary table data.

### Recommended Layout

This generic layout keeps server and client artifacts separate:

```text
config/excel/                    # workbook sources: *.xlsx
build/excel/server/proto/        # server build intermediates: *.proto + *.protoset
build/excel/client/proto/        # client build intermediates: *.proto + *.protoset
server/gen/excel/                # *.pb.go, *.structure.go, *.excel.go, tables.go
server/res/excel/                # runtime *Table.json or *Table.bin
client/addons/proto/             # tools/protoc-gen-gdscript/godot runtime
client/addons/excel/             # tools/protoc-gen-gdscript-excel/godot runtime
client/script/gen/excel/         # *.pb.gd, *.excel.gd, tables.gd
client/excel/                    # *Table.bin, or *Table.bin.idx + *Table.bin.chk_*
```

Use separate proto directories for server and client targets. `--targets` can remove fields, while the chosen index representation and GDScript array storage may also differ.

### Workbook Format

[`tools/excelc/examples/ExampleCN.xlsx`](./tools/excelc/examples/ExampleCN.xlsx) and [`ExampleEN.xlsx`](./tools/excelc/examples/ExampleEN.xlsx) are complete samples.

Both workbooks use the same layout and demonstrate object and enum declarations, field aliases, scalar types, lists, object lists, maps, string escaping, `scope=c` / `scope=s`, single-column and composite `unique_index` / `index` definitions, multiple data pages, and note pages. Each data row fills only the fields needed by that example; omitted fields retain their Protobuf defaults.

#### Workbook Contents

- One workbook represents one logical table.
- The optional `@types` sheet declares reusable objects and enums for that workbook.
- Ordinary sheets whose names start with a letter, including Chinese characters, are exported and merged in tab order; prefix note sheets excluded from data export with `#`.

#### The `@types` Sheet

Row 1 contains column names. Starting at row 2, each row declares one object field or enum value. A type can span multiple rows. A blank `EnumValue` creates an object field, while a populated value creates an enum value; one type cannot mix both forms.

| Column                                 | Description                                                                                                      |
|----------------------------------------|------------------------------------------------------------------------------------------------------------------|
| `ObjectType` / `Type`                  | Type name. Reuse it for all fields of one object or all values of one enum.                                      |
| `FieldName`                            | Object field name or enum value name.                                                                            |
| `FieldType`                            | Object field type; accepts built-ins, declared types, and `Type[]` arrays. Unused for enum values.               |
| `EnumValue` / `Value`                  | Leave blank for an object field, or provide a nonnegative integer for an enum value. A proto3 enum starts at `0`. |
| `Alias`                                | Optional object-field or enum-value alias accepted in data cells; Chinese text is allowed.                       |
| `Default`                              | Reserved; it currently does not affect schema generation or data export.                                         |
| `Meta`                                 | Supports `separator`, `scope`, and `pb_field_number`; index options only apply to data-page fields.              |
| `Comment`                              | Written to the generated Protobuf declaration.                                                                   |

Names emitted as Protobuf identifiers, including type names, object fields, enum values, and data-page fields, must match `[A-Za-z][A-Za-z0-9_]*`; validated names are then converted to UpperCamelCase. Aliases may contain Chinese text but cannot contain ASCII spaces, YAML indicator characters (`-?:,[]{}#&*!|>'"%@`), a backtick, a backslash, or Unicode control characters.

#### Data Pages

The first data page defines fields, types, `Meta`, comments, and indexes. Later pages use only field names in row 1 to bind data. Rows 2 through 4 are ignored on later pages but must remain as placeholders so that data starts at row 5:

| Row      | Contents      |
|----------|---------------|
| 1        | Field name.   |
| 2        | Field type.   |
| 3        | Field `Meta`. |
| 4        | Comment.      |
| 5 onward | Table data.   |

Declaration cells are trimmed, with CRLF and CR normalized to LF. Control characters in comments are converted to visible escapes such as `\n` and `\t`.

- The first blank field-name cell in row 1 ends the active column range. That column and every column to its right are ignored even when other rows contain values.
- Before that boundary, a column whose converted field name does not start with a letter is also ignored. A `#`-prefixed column is useful for row comments and does not terminate active columns to its right.
- Starting at row 5, a row is skipped without consuming a row offset when every recognized field column is blank. A blank row never terminates the page; later nonblank rows are still exported. Content only in comment columns or beyond the active-column boundary does not make a row nonblank.
- A row is exported when any recognized field column is nonblank, and omitted fields retain their Protobuf defaults. `--targets` filtering does not change the blank-row test, keeping row offsets aligned between targets.
- Later pages bind columns by field name. Columns may be reordered or omitted; an omitted field is treated as an empty cell and retains its Protobuf default. A later page cannot introduce a field absent from the first page or repeat a field name.
- Merged pages share continuous row offsets and indexes. Unique indexes detect duplicates across pages, and non-unique index results preserve page and source-row order.

#### Cell Syntax

| Type            | Example and behavior                                                                                   |
|-----------------|--------------------------------------------------------------------------------------------------------|
| Scalar          | `1`, `3.14`, `true`, or `HelloWorld`.                                                                          |
| `bytes`         | Base64 text such as `SGVsbG9Xb3JsZA==`.                                                                        |
| Enum            | Numeric value, enum name, or configured alias.                                                                 |
| Repeated scalar | Comma-separated by default, such as `1,2,3`; YAML sequences such as `[1, 2, 3]` are also supported.           |
| Object          | YAML-style mapping such as `id: 1, name: Example, tags: [1, 2]`; field names and aliases are accepted.       |
| Repeated object | Use a YAML sequence such as `[{ID: 1}, {ID: 2}]`, or split mapping fragments with the field `separator`, such as `A: 1, B: Hello \| A: 2, B: World`. |
| Map             | YAML-style mappings only, such as `{1: Alpha, 2: Beta}` or `{1: {id: 1, name: Alpha}}`; `separator` is not used for maps. |

In an object mapping, write each field and value as `field: value`, with a space after the colon:

```yaml
name: Example  # Valid
name:Example   # Invalid; excelc asks for a space after the colon
```

This check applies to objects, repeated objects, and object values inside maps. It does not inspect ordinary strings or map keys. Object fields absent from the current schema remain ignored, allowing the same data to work with schemas trimmed by `targets`/`scope`.

A YAML mapping cannot contain duplicate keys. This is checked recursively in objects, repeated objects, and maps. For example, `{1: Alpha, 1: Beta}` reports the first and duplicate occurrences of key `1` instead of silently selecting one value. The reported line and column are positions within the YAML text in the current Excel cell, not worksheet coordinates; the surrounding error identifies the worksheet, data row, and field name.

#### Value Parsing Rules

- Field parsing trims leading and trailing cell whitespace first; a value that becomes empty is ignored. Quote values when leading or trailing whitespace must be preserved, for example `"  Hello  "`.
- A line break entered directly in an Excel data cell is stored in Protobuf as a line break and displays as such in Go and Godot; platform-specific line endings are normalized to LF, while line breaks at the beginning or end of a cell are trimmed as surrounding whitespace. A data-page column of type `string` does not treat an unquoted `\n` as a line break: `first\nsecond` is preserved literally, `\n` in `"first\nsecond"` is converted to a line break by YAML double-quote escaping, and `\n` in `'first\nsecond'` remains literal.
- Each repeated field selects its parsing mode independently. A YAML sequence root uses standard YAML sequence parsing, including `[1, 2, 3]` and block-style `- item`; any other root uses that field's own `separator`, which defaults to `,`.
- A separator is recognized only outside quotes and outside nested `{}` or `[]`. Each repeated-object fragment is parsed as an independent YAML mapping, and its outer `{}` may be omitted.
- The default separator `,` also separates object fields, so it is ambiguous for multi-field repeated objects whose outer `{}` are omitted. For example, `A: 1, B: Hello, A: 2, B: World` is ambiguous; use another object-list separator such as `|`, keep the outer `{}`, or use a standard YAML sequence.
- Quotes can protect a nested repeated field's separators from the outer split. After the containing object is parsed, the nested field still applies its own separator. If both the outer field and `D` use `|`, `A: 1, B: HelloWorld, D: "1|2|3" | A: 2, B: HAHAHAHAHA, D: "4|5|6"` produces two objects whose `D` fields each contain three items.
- Elements of a standard YAML sequence are not split again. For a repeated string, `D: ["1|2|3"]` means one string item, while `D: "1|2|3"` enters separator mode and produces three items.
- Single and double quotes are YAML syntax and are removed from assigned string values; escapes inside quotes follow YAML rules. If any list item fails to parse, the complete field fails without writing a partial result.

#### Field Metadata

Metadata uses query-string syntax, for example `scope=client&sorted_unique_index=1`:

| Parameter             | Description                                                                                                          |
|-----------------------|----------------------------------------------------------------------------------------------------------------------|
| `scope`               | Repeatable target label used with `--targets`; fields without a scope are visible to every target.                   |
| `separator`           | Custom separator for repeated-field cells; defaults to `,` and may contain multiple characters. It cannot be empty or contain whitespace (including at either end), control characters, or `: ' " \\ { } [ ]`. Map fields always use YAML mapping syntax. |
| `pb_field_number`     | Overrides the Protobuf field number. It must be positive, outside the reserved range, and unique within its message. |
| `unique_index`        | Logical unique-index group whose representation is selected by `--pb_unique_index_as`.                               |
| `hash_unique_index`   | Forces a hash-based unique index.                                                                                    |
| `sorted_unique_index` | Forces a sorted unique index.                                                                                        |
| `index`               | Logical non-unique group whose representation is selected by `--pb_index_as`.                                        |
| `hash_index`          | Forces a hash-based non-unique index.                                                                                |
| `sorted_index`        | Forces a sorted non-unique index.                                                                                    |

### Index Model

| Index                 | Main representation                           | Lookup behavior                                        | Typical use                               |
|-----------------------|-----------------------------------------------|--------------------------------------------------------|-------------------------------------------|
| `hash_unique_index`   | `hash -> row offset`, plus a collision bucket | Average constant-time lookup, one matching row         | Query-heavy servers with enough memory.   |
| `sorted_unique_index` | `Values + Offsets`                            | Binary search over `Values`                            | Clients that want to avoid map objects.   |
| `hash_index`          | `hash -> Offsets` bucket                      | Average constant-time lookup, multiple rows per key    | Query speed over bucket/list memory cost. |
| `sorted_index`        | `Values + Starts + Offsets`                   | Binary-search key, then read a contiguous offset range | Memory-sensitive client devices.          |

Index configuration rules:

- A single-column index assigns a tag to one field; a composite index reuses the same tag on multiple fields.
- One field may participate in multiple indexes, for example `hash_unique_index=1&hash_unique_index=2`.
- The four physical index types keep independent tag groups and may reuse numbers. Do not assign the same field and tag to multiple index types.
- Unique lookups return one row; non-unique lookups return rows in original row-offset order.
- Lookup results reference objects in the table message's `Rows`; they are not cloned.
- `unique_index` defaults to `hash_unique_index`, while `index` defaults to `sorted_index`; command-line options can override both defaults.

### Complete Workflow

The following POSIX shell commands use neutral paths and names. They emit hash indexes for the server and sorted indexes for the client. The `server` / `client` target labels are conventions and can be replaced. Use equivalent line continuations and per-file loops in Windows batch or PowerShell.

#### 1. Generate Server and Client Schemas

```bash
excelc proto \
  --excel_files=./config/excel/Config.xlsx \
  --pb_out=./build/excel/server/proto \
  --pb_package=excel \
  --pb_options='[go_package=./excel]' \
  --pb_unique_index_as=hash_unique_index \
  --pb_index_as=hash_index \
  --targets=server

excelc proto \
  --excel_files=./config/excel/Config.xlsx \
  --pb_out=./build/excel/client/proto \
  --pb_package=excel \
  --pb_options='[go_package=./excel]' \
  --pb_unique_index_as=sorted_unique_index \
  --pb_index_as=sorted_index \
  --gdscript_index_array=packed_int64 \
  --targets=client
```

To process every workbook in a directory, replace `--excel_files` with `--excel_dir=./config/excel`.

#### 2. Generate Server Code and Descriptor Sets

This is a POSIX shell example. `PROTOBUF_INCLUDE` must point to the include root containing `google/protobuf/descriptor.proto`.

```bash
SERVER_PROTO_DIR=./build/excel/server/proto
PROTOBUF_INCLUDE=./third_party/protobuf/include

for proto in "$SERVER_PROTO_DIR"/*.proto; do
  protoc \
    -I"$SERVER_PROTO_DIR" \
    -I"$PROTOBUF_INCLUDE" \
    --include_imports \
    --retain_options \
    --descriptor_set_out="${proto}set" \
    --go_out=./server/gen \
    --go-structure_out=./server/gen \
    --go-excel_out=./server/gen \
    "$proto" || exit 1
done

excelc code \
  --pb_dir="$SERVER_PROTO_DIR" \
  --pb_package=excel \
  --go_out=./server/gen/excel
```

`"${proto}set"` turns `Config.proto` into `Config.protoset` and the dependency schema `excelc.proto` into `excelc.protoset`. Do not merge everything into an arbitrarily named descriptor set: `excelc` resolves files by workbook name.

#### 3. Generate Client Code and Descriptor Sets

```bash
CLIENT_PROTO_DIR=./build/excel/client/proto
PROTOBUF_INCLUDE=./third_party/protobuf/include

for proto in "$CLIENT_PROTO_DIR"/*.proto; do
  protoc \
    -I"$CLIENT_PROTO_DIR" \
    -I"$PROTOBUF_INCLUDE" \
    --include_imports \
    --retain_options \
    --descriptor_set_out="${proto}set" \
    --gdscript_out=./client/script/gen \
    --gdscript_opt=string_as_string_name=true \
    --gdscript-excel_out=./client/script/gen \
    --gdscript-excel_opt=string_as_string_name=true \
    "$proto" || exit 1
done

excelc code \
  --pb_dir="$CLIENT_PROTO_DIR" \
  --pb_package=excel \
  --gdscript_out=./client/script/gen/excel \
  --gdscript_default_data_dir=res://excel/ \
  --gdscript_autoload=false
```

Keep `string_as_string_name` identical for `protoc-gen-gdscript` and `protoc-gen-gdscript-excel`. Each `*.excel.gd`, its `*.pb.gd`, and the aggregate `tables.gd` should share the same output layer.

#### 4. Export Data

A server can export readable JSON for inspection and hot loading:

```bash
excelc data \
  --excel_files=./config/excel/Config.xlsx \
  --pb_dir=./build/excel/server/proto \
  --pb_package=excel \
  --targets=server \
  --json_out=./server/res/excel \
  --json_multiline=true \
  --json_indent="  "
```

A client can export chunked binary data for on-demand loading:

```bash
excelc data \
  --excel_files=./config/excel/Config.xlsx \
  --pb_dir=./build/excel/client/proto \
  --pb_package=excel \
  --targets=client \
  --binary_out=./client/excel \
  --binary_chunked=true \
  --binary_chunk_size=256
```

Chunked mode emits `*.bin.idx` and `*.bin.chk_*`. Indexes and the chunk manifest live in `.idx`; rows live in chunk files, and lookup loads only the chunk containing the matched row. `--binary_chunk_size` is a row count, not a byte count, and should be tuned for row size and access patterns.

#### 5. Automation Recommendations

- Split “schema/code compilation” from “data export.” Header, type, metadata, scope, or index changes require recompilation; ordinary row edits only require `excelc data`.
- Remove stale `.proto`, `.protoset`, and generated code before recompiling so descriptors from deleted workbooks are not scanned by `excelc code`.
- When generating inside a Godot project, preserve `.uid` files that still have matching scripts and remove only orphaned `.uid` files to avoid unnecessary resource UID churn.
- Check every command's exit code and stop immediately instead of exporting new data with an old descriptor set.

### `excelc`

`excelc` provides three subcommands. Run `excelc <command> --help` for the complete flag list.

#### `excelc proto`

Generates `excelc.proto` and one table schema `.proto` for each workbook.

| Option                   | Description                                                                                                 |
|--------------------------|-------------------------------------------------------------------------------------------------------------|
| `--excel_files`          | Explicit input files; preferred over `--excel_dir`.                                                         |
| `--excel_dir`            | Scans a directory for Excel files.                                                                          |
| `--pb_out`               | `.proto` output directory.                                                                                  |
| `--pb_package`           | Proto package; defaults to `excel`.                                                                         |
| `--pb_options`           | File options written into generated proto files, such as `go_package`.                                      |
| `--pb_imports`           | Additional imports written into generated proto files.                                                      |
| `--pb_custom_options`    | Base number for Excel custom options; defaults to `10000` and must stay consistent across one artifact set. |
| `--targets`              | Target labels used with field `scope` to trim the schema.                                                   |
| `--pb_unique_index_as`   | Default `unique_index` representation: `hash_unique_index` or `sorted_unique_index`.                        |
| `--pb_index_as`          | Default `index` representation: `hash_index` or `sorted_index`.                                             |
| `--gdscript_index_array` | Uses `packed_int64` or `array` for GDScript index vectors; defaults to `packed_int64`.                      |

#### `excelc code`

Reads `*.protoset` files under `--pb_dir` and generates aggregate loaders:

- `--go_out` emits `tables.go` with `Tables`, `LoadBinaryFiles`, and `LoadJsonFiles`.
- `--gdscript_out` emits `tables.gd`, exports tables/messages/enums, and loads ordinary or chunked binary data.
- `--gdscript_class_name` controls the aggregate script's `class_name`; the default is `Tables`, and an empty value disables it.
- `--gdscript_default_data_dir` defaults to `res://excel/`.
- `--gdscript_autoload` defaults to `true` and calls `load_data()` from `_ready()`. Set it to `false` when startup ordering or threading is managed by the application.

#### `excelc data`

Reads workbooks and matching `*.protoset` files, builds dynamic table messages, and exports data:

| Option                               | Description                                                           |
|--------------------------------------|-----------------------------------------------------------------------|
| `--excel_files` / `--excel_dir`      | Input workbooks.                                                      |
| `--pb_dir` / `--pb_package`          | Descriptor-set directory and proto package.                           |
| `--targets`                          | Must match the target used to generate the selected schema directory. |
| `--json_out`                         | Emits `*.json`.                                                       |
| `--json_multiline` / `--json_indent` | Controls readable JSON formatting.                                    |
| `--binary_out`                       | Emits `*.bin`.                                                        |
| `--binary_chunked`                   | Switches to `.bin.idx + .bin.chk_*`.                                  |
| `--binary_chunk_size`                | Maximum rows per chunk; defaults to `10000`.                          |

### `protoc-gen-go-excel`

This plugin only targets schemas produced by `excelc proto` and emits `*.excel.go`. It reads table/index custom options and adds:

- Unique indexes: `LookupBy...` returns `(*Row, bool)`, while `GetBy...` panics on a miss. The first unique index also gets `Lookup` / `Get` aliases.
- Non-unique indexes: `LookupBy...` returns `[]*Row`; a miss returns `nil`, which can be used as an empty slice. No `Get` method is generated.
- Single- and multi-column indexes, hash-collision verification, and hash/sorted representations.

The plugin has no custom options. Generated code depends on [`tools/excelc/excelutils`](./tools/excelc/excelutils), so the application Go module must depend on this repository.

### `protoc-gen-gdscript-excel`

This plugin only targets schemas produced by `excelc proto` and emits `*.excel.gd` with ordinary and chunked table wrappers:

- `rows`, `row_count`, `row_at`, and their async counterparts.
- Unique `lookup_by_<index_type>_<fields>` methods returning a row or `null`.
- Non-unique methods with the same `lookup_by_...` naming and an `Array[Row]` result.
- `lookup` / `lookup_async` aliases for the first unique index.
- On-demand chunk loading from async methods on chunked wrappers.

| Option                  | Default | Description                                                                      |
|-------------------------|---------|----------------------------------------------------------------------------------|
| `string_as_string_name` | `false` | Must match the setting used by `protoc-gen-gdscript` in the same generation run. |

Generated code requires both [`tools/protoc-gen-gdscript/godot`](./tools/protoc-gen-gdscript/godot) and [`tools/protoc-gen-gdscript-excel/godot`](./tools/protoc-gen-gdscript-excel/godot). Lookup results reference objects in `Rows`; they are not cloned.

### GDScript Index Arrays

`--gdscript_index_array=packed_int64` maps internal index `Values`, `Starts`, and `Offsets` vectors to `PackedInt64Array`, reducing Variant container overhead and improving sequential access. Use `array` for compatibility with older runtimes that expect `Array[int]`.

This option:

- Applies only to internal Excel index messages, not ordinary repeated fields.
- Does not change the Protobuf wire format.
- Does not affect Go code or server data structures.
- Requires rerunning `excelc proto` and regenerating both `*.pb.gd` and `*.excel.gd` after a change.

## Property Synchronization Generation

### `propc`

`propc` scans `+prop-sync-gen:` annotations immediately above types and methods in one Go declaration file, then emits an adjacent `*.sync.gen.go`. A typical declaration is:

```go
package state

import (
	pb "example.com/project/server/gen/pb"
	"git.golaxy.org/scaffold/addins/propview"
)

//go:generate propc

// +prop-sync-gen:sync=true
type ProfileProp struct {
	propview.PropT[*pb.Profile]
}

// +prop-sync-gen:sync=true
func (p *ProfileProp) SetName(name string) {
	p.State().Name = name
}
```

Run:

```bash
go generate ./...
```

Or specify the declaration file directly:

```bash
propc --decl_file=profile_prop.go
```

The generated `profile_prop.sync.gen.go` defines `ProfilePropSync` and wraps `Load`, `Save`, `Managed`, and annotated operations. Each synchronized operation invokes the original implementation, increments its revision, and broadcasts the operation through `propview`.

Notes:

- The annotation must be immediately above the target type or method.
- Only types with `sync=true` receive wrappers.
- Only pointer-receiver methods on a selected type can be synchronized.
- The underlying state is normally a message implementing GAP `variant.Value`, which can be generated with `protoc-gen-go-variant`.
- `//go:generate propc` uses the `GOFILE` environment variable supplied by Go; use `--decl_file` for manual invocation.

## Runtime Components

### Go Add-ins

#### `addins/propview`

`propview` provides managed property tables, serialization, revisions, cross-service loading/persistence, and incremental synchronization. It is normally used with `*Sync` types emitted by `propc` and installed into a Golaxy runtime through `propview.AddIn`.

#### `addins/goscr`

`goscr` is a Yaegi-based service-level script add-in. It can load one or more local or remote script projects and integrate scripted entities/components with the Golaxy lifecycle. `addins/goscr/dynamic` manages projects, solutions, and hot reloads; `addins/goscr/fwlib` contains symbols exported into the script environment.

### Godot Runtime Directories

| Directory                               | Required when                                                                       |
|-----------------------------------------|-------------------------------------------------------------------------------------|
| `tools/protoc-gen-gdscript/godot`       | Any generated `*.pb.gd` is used.                                                    |
| `tools/protoc-gen-gdscript-excel/godot` | Any generated `*.excel.gd` is used; the previous runtime is still required.         |
| `godot/rpcli`                           | The Godot client connects through GAP/GTP or generation enables `gap_variant=true`. |
| `godot/resty`                           | The Godot client uses regular HTTP, downloads, or SSE.                              |

The directories do not require fixed installation paths. A common layout is:

```text
res://addons/proto/
res://addons/excel/
res://addons/rpcli/
res://addons/resty/
res://script/gen/proto/
res://script/gen/excel/
res://excel/
```

The generated Excel aggregate script can be registered as an autoload:

```ini
[autoload]

Excel="*res://script/gen/excel/tables.gd"
```

When generation used `--gdscript_autoload=false`, call it explicitly during startup:

```gdscript
if !Excel.load_data("res://excel/"):
	push_error("failed to load excel data")
```

Register the RPC client as an autoload and connect:

```ini
[autoload]

RPCli="*res://addons/rpcli/golaxy_rpcli.gd"
```

```gdscript
var ok := await RPCli.connect_to_async(
	"ws://127.0.0.1:8080",
	GolaxyClient.PROTOCOL_WEBSOCKET,
	"user_id",
	"token"
)
```

Register the HTTP client as an autoload and create an isolated request snapshot:

```ini
[autoload]

Resty="*res://addons/resty/resty_client.gd"
```

```gdscript
var response := await (
	Resty.set_base_url("https://api.example.com")
	.r()
	.set_bearer_auth("token")
	.set_query_param("page", 1)
	.get_async("/users")
)
```

`Resty` also supports JSON/form/raw bodies, path parameters, output files, concurrent request handles, and `Resty.sse(url)`.

## Repository Layout

| Path                                                                   | Responsibility                                            |
|------------------------------------------------------------------------|-----------------------------------------------------------|
| [`addins/goscr`](./addins/goscr)                                       | Go scripting add-in, dynamic projects, and hot reloads.   |
| [`addins/propview`](./addins/propview)                                 | Managed properties and cross-endpoint synchronization.    |
| [`tools/excelc`](./tools/excelc)                                       | Excel schema, code, and data generation CLI.              |
| [`tools/excelc/examples`](./tools/excelc/examples)                     | Sample Excel workbooks.                                   |
| [`tools/excelc/excelutils`](./tools/excelc/excelutils)                 | Go table loading, index, hashing, and comparison helpers. |
| [`tools/propc`](./tools/propc)                                         | Property synchronization generator.                       |
| [`tools/protoc-gen-go-structure`](./tools/protoc-gen-go-structure)     | Go Protobuf deep-copy plugin.                             |
| [`tools/protoc-gen-go-variant`](./tools/protoc-gen-go-variant)         | Go GAP variant plugin.                                    |
| [`tools/protoc-gen-go-excel`](./tools/protoc-gen-go-excel)             | Go Excel lookup plugin.                                   |
| [`tools/protoc-gen-gdscript`](./tools/protoc-gen-gdscript)             | GDScript Protobuf plugin and runtime.                     |
| [`tools/protoc-gen-gdscript-excel`](./tools/protoc-gen-gdscript-excel) | GDScript Excel plugin and runtime.                        |
| [`godot/rpcli`](./godot/rpcli)                                         | Godot GAP/GTP RPC client.                                 |
| [`godot/resty`](./godot/resty)                                         | Godot HTTP/SSE client.                                    |

## Related Repositories

- [Golaxy Distributed Service Development Framework Core](https://github.com/pangdogs/core)
- [Golaxy Distributed Service Development Framework](https://github.com/pangdogs/framework)

## License

This project is licensed under the [GNU Lesser General Public License v2.1](./LICENSE).
