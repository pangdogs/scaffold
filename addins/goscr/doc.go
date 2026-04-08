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

// Package goscr provides the service-side Go scripting add-in for scaffold.
/*
Package goscr 基于 yaegi 为框架服务提供 Go 脚本装载、脚本实体原型声明、
脚本组件桥接与热更新能力。

它通常与 goscr/dynamic 和 goscr/fwlib 配合使用：前者负责脚本工程与方案管理，
后者向解释器导出 core、framework 与 scaffold 的可用符号。
*/
package goscr
