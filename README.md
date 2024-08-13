# Scaffold
[English](./README.md) | [简体中文](./README.zh_CN.md)

## Introduction
Based on the [**Golaxy Distributed Service Development Framework**](https://github.com/pangdogs/framework), this project provides scaffolding functions and tools for game development. It includes tools for table processing (`Excel Compile And Export`), Protobuf plugins (`Protobuf Plugins`), and a property synchronization system (`Property Synchronous`), among others.

## Tools Directory
| Directory | Description |
| --------- | ----------- |
| [/tools/excelc](https://github.com/pangdogs/scaffold/tools/excelc) | Excel table processing tool, based on Protobuf technology, allows exporting the structure and data configured in Excel tables. |
| [/tools/propc](https://github.com/pangdogs/scaffold/tools/propc) | Property synchronization code generation tool, capable of generating code for property synchronization operations. |
| [/tools/protoc-gen-go-excel](https://github.com/pangdogs/scaffold/tools/protoc-gen-go-excel) | Protobuf plugin that adds some table operation-related functions to the Golang code generated from the Protobuf structure exported from Excel tables. |
| [/tools/protoc-gen-go-safe](https://github.com/pangdogs/scaffold/tools/protoc-gen-go-safe) | Protobuf plugin that adds some auxiliary functions when generating Golang code. |
| [/tools/protoc-gen-go-variant](https://github.com/pangdogs/scaffold/tools/protoc-gen-go-variant) | Protobuf plugin that adds functions related to variant types when generating Golang code, allowing the generated Protobuf structure to support the Golaxy framework's RPC system. |

## Plugins Directory
| Directory | Description |
| --------- | ----------- |
| [/plugins/acl](https://github.com/pangdogs/scaffold/plugins/acl) | Access Control List (`ACL`) plugin, generally used for the Login service. |
| [/plugins/view](https://github.com/pangdogs/scaffold/plugins/view) | Property View plugin, used to support the synchronization of entity properties. |
