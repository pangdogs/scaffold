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
### Excel table pipeline
1. Author `.xlsx` tables.
2. Run `excelc proto` to generate table-oriented `.proto` files.
3. Run `protoc` with `protoc-gen-go-excel`, `protoc-gen-go-structure`,
   `protoc-gen-go-variant`, or the GDScript plugins as needed.
4. Run `excelc data` to export binary and/or JSON table data.

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
| [`./tools/protoc-gen-gdscript-excel`](./tools/protoc-gen-gdscript-excel) | GDScript protobuf plugin for Excel table wrappers |

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
