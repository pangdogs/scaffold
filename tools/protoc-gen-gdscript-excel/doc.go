/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

// Package main implements the protoc-gen-gdscript-excel protobuf plugin.
/*
Package main implements protoc-gen-gdscript-excel, a protoc plugin that emits
Godot-facing `*.excel.gd` table wrappers and index lookup helpers for proto
schemas generated from Excel tables.

Generated Excel wrappers depend on two runtime layers:

1. `tools/protoc-gen-gdscript/libs`, because the wrappers call protobuf helper
   classes such as `ProtoUtils` and `ProtoInputFile`.
2. `tools/protoc-gen-gdscript-excel/libs`, because chunk loading, index
   conversion, and binary-search helpers live in that runtime directory.

These runtime scripts do not need fixed directory names. In real projects it
is common to place both runtime layers into one shared location such as
`libs` or `addons/<name>`.

Keep each generated `*.excel.gd` file next to its matching `*.pb.gd` file.
The wrapper preloads `./<name>.pb.gd`, and the protobuf file layout should also
preserve the relative structure of the source `.proto` set.

Package main 实现 protoc-gen-gdscript-excel 插件，为 Excel 配表导出的
proto 结构生成 Godot 侧的 `*.excel.gd` 表包装器和按索引查询辅助方法。

生成的 Excel 包装器依赖两层运行时脚本：

1. `tools/protoc-gen-gdscript/libs`，因为包装器会调用 `ProtoUtils`、
   `ProtoInputFile` 等 protobuf 基础能力。
2. `tools/protoc-gen-gdscript-excel/libs`，因为分块加载、索引值转换和
   二分查找等辅助逻辑都位于这个运行时目录中。

这些运行时脚本不要求放在固定目录名下。实际项目里通常会把两层运行时都统一
放到 `libs` 或 `addons/<name>` 之类的共享目录中。

每个生成的 `*.excel.gd` 都应与对应的 `*.pb.gd` 放在同一个输出目录下。
包装器会预加载 `./<name>.pb.gd`，而 `*.pb.gd` 自身也应尽量保持源
`.proto` 集合的相对目录结构，以保证跨文件引用可正常解析。
*/
package main
