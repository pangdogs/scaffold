# Scaffold
[English](./README.md) | [简体中文](./README.zh_CN.md)

## 简介
基于[**Golaxy分布式服务开发框架**](https://github.com/pangdogs/framework)，为开发游戏项目提供一些脚手架功能和工具，包含表格处理工具（`Excel Compile And Export`）、Protobuf插件（`Protobuf Plugins`）和属性同步系统（`Property Synchronous`）等。

## 工具目录
| Directory | Description                                                                     |
| --------- |---------------------------------------------------------------------------------|
| [/tools/excelc](https://github.com/pangdogs/scaffold/tree/main/tools/excelc) | Excel表格处理工具，基于Protobuf技术，可以导出Excel表配置的结构和数据。                                    |
| [/tools/propc](https://github.com/pangdogs/scaffold/tree/main/tools/propc) | 属性同步代码生成工具，可以生成属性同步操作代码。                                                        |
| [/tools/protoc-gen-go-excel](https://github.com/pangdogs/scaffold/tree/main/tools/protoc-gen-go-excel) | Protobuf插件，使Excel表导出的Protobuf结构，生成Golang代码时，加入一些表格操作相关函数。                       |
| [/tools/protoc-gen-go-safe](https://github.com/pangdogs/scaffold/tree/main/tools/protoc-gen-go-safe) | Protobuf插件，生成Golang代码时，加入一些辅助函数。                                                |
| [/tools/protoc-gen-go-variant](https://github.com/pangdogs/scaffold/tree/main/tools/protoc-gen-go-variant) | Protobuf插件，生成Golang代码时，加入一些可变类型（Variant）相关函数，使Golaxy框架的RPC系统中能够支持生成出的Protobuf结构。 |

## 插件目录
| Directory                                                                    | Description                                              |
|------------------------------------------------------------------------------|----------------------------------------------------------|
| [/addins/acl](https://github.com/pangdogs/scaffold/tree/main/addins/acl)     | 访问控制表（`Access Control List`），一般用于Login服务。              |
| [/addins/goscr](https://github.com/pangdogs/scaffold/tree/main/addins/goscr) | Golang脚本化（`Golang Script`），支持解释执行Golang代码，用于支持逻辑层代码热修复。 |
| [/addins/view](https://github.com/pangdogs/scaffold/tree/main/addins/view)   | 属性视图（`Property View`），用于支持同步实体属性。                      |

## 安装
```
go get -u git.golaxy.org/scaffold
```

## 相关项目
- [Golaxy分布式服务开发框架内核](https://github.com/pangdogs/core)
- [Golaxy分布式服务开发框架](https://github.com/pangdogs/framework)