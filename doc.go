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

// Package scaffold provides the tooling and auxiliary add-ins used around
// Golaxy-based game and realtime service projects.
/*
Package scaffold 在 git.golaxy.org/core 与 git.golaxy.org/framework 之外补齐
项目脚手架能力，包括：

  - Excel 配表编译、表结构 proto 生成，以及 Go/GDScript 访问代码导出；
  - 面向 GAP variant、结构拷贝和表查询访问的 Protobuf 插件；
  - 基于 yaegi 的 Go 脚本加载、脚本实体声明与热更新 add-in；
  - 面向实体状态复制的属性视图与属性同步代码生成。

典型用法是先用 tools/excelc、tools/propc 和各类 protoc 插件生成静态代码与数据，
再在 framework 服务和 runtime 中装配 goscr、propview 等 add-in，以补齐热更新、
配表访问和跨服务属性同步能力。
*/
package scaffold
