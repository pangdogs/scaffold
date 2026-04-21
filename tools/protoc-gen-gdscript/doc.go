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

// Package main implements the protoc-gen-gdscript protobuf plugin.
/*
Package main implements protoc-gen-gdscript, a protoc plugin that emits
Godot-facing `*.pb.gd` files for protobuf enums, messages, serialization,
deserialization, cloning, hashing, and equality checks.

The generated scripts are not self-contained. After generation, copy every
script in `tools/protoc-gen-gdscript/libs` into your Godot project so classes
such as `ProtoMessage`, `ProtoUtils`, `ProtoInputStream`, `ProtoOutputStream`,
`ProtoInputFile`, and `ProtoOutputBuffer` are registered through `class_name`
and can be resolved by generated code.

These runtime scripts do not need a fixed directory name. In real projects it
is common to place them in one shared location such as `libs` or
`addons/<name>`.

Keep generated `*.pb.gd` files in the same relative layout as the source
`.proto` files. Cross-file references are emitted as relative `preload(...)`
calls, so moving one generated file without the others can break imports.

Package main 实现 protoc-gen-gdscript 插件，把 protobuf 定义转换为 Godot
侧可用的 `*.pb.gd` 文件，包含枚举、消息类型、序列化、反序列化、克隆、哈希
与值比较逻辑。

生成结果并不是自包含的。生成后需要把
`tools/protoc-gen-gdscript/libs` 里的脚本一并拷贝到 Godot 项目中，让
`ProtoMessage`、`ProtoUtils`、`ProtoInputStream`、
`ProtoOutputStream`、`ProtoInputFile`、`ProtoOutputBuffer` 等通过
`class_name` 注册的运行时类型可被生成代码直接解析。

这些运行时脚本不要求放在固定目录名下。实际项目里通常会把它们统一放到
`libs` 或 `addons/<name>` 之类的共享目录中。

生成的 `*.pb.gd` 文件应尽量保持与源 `.proto` 相同的相对目录结构。跨文件
类型引用会被生成为相对 `preload(...)`，单独移动某个生成文件可能导致导入
路径失效。
*/
package main
