# Scaffold
[English](./README.md) | [简体中文](./README.zh_CN.md)

## 简介
`scaffold` 是
[**Golaxy 分布式服务开发框架**](https://github.com/pangdogs/framework)
周边的脚手架与工具仓库，关注运行时之外的项目搭建能力，包括 Excel
配表编译、Protobuf/GDScript 代码生成、Go 脚本热更新，以及实体属性同步。

仓库当前主要分为两层：

- `addins`：挂接到 `git.golaxy.org/framework` 的服务级或运行时级扩展。
- `tools`：在构建期使用的命令行生成器和 Protobuf 插件。

## 提供的能力
- `addins/goscr`：基于 Yaegi 的服务级 Go 脚本 add-in，支持脚本实体/组件声明、脚本工程加载，以及本地或远端脚本热更新。
- `addins/propview`：运行时级属性视图 add-in，负责实体属性的托管、加载、保存和跨服务/客户端同步。
- `tools/propc`：属性同步代码生成器，扫描带注解的 Go 声明并生成基于 `propview` 的同步包装代码。
- `tools/excelc`：把 `.xlsx` 表格转换为 Proto 结构、访问代码，以及二进制/JSON 数据文件的 Excel 工具链。
- `tools/protoc-gen-go-excel`：为生成的 Go Protobuf 代码补充表查询与索引访问函数。
- `tools/protoc-gen-go-structure`：为 Protobuf 消息补充深拷贝辅助函数，覆盖消息、切片、映射、字节数组等字段。
- `tools/protoc-gen-go-variant`：让生成的 Protobuf 消息能够作为 Golaxy GAP/RPC 栈中的 variant 值使用。
- `tools/protoc-gen-gdscript`：生成 Godot 侧可用的 GDScript Protobuf 消息类型与序列化逻辑。
- `tools/protoc-gen-gdscript-excel`：为 Excel 生成的表结构补充 GDScript 表包装器与索引查询函数。

## 典型工作流
### Excel 配表流水线
1. 编写 `.xlsx` 配表文件。
2. 使用 `excelc proto` 生成面向表结构的 `.proto` 文件。
3. 按需用 `protoc` 搭配 `protoc-gen-go-excel`、`protoc-gen-go-structure`、`protoc-gen-go-variant` 或 GDScript 插件生成代码。
4. 使用 `excelc data` 导出二进制和/或 JSON 数据文件。

### 属性同步
1. 在 Go 中定义属性状态类型和需要同步的方法。
2. 按 `tools/propc` 的约定添加注解并生成 `*.sync.gen.go`。
3. 通过 `addins/propview` 声明属性，让其具备加载、保存和按 revision 复制的能力。

### 脚本热更新
1. 在服务侧安装 `addins/goscr`。
2. 通过 `goscr.With.Projects(...)` 指定一个或多个本地或远端脚本工程。
3. 用 `goscr.BuildEntityPT(...)` 结合脚本元信息声明实体原型。
4. 由 add-in 自动监听本地文件变更或远端源码变化并重新加载脚本方案。

## Godot 运行时库
- `tools/protoc-gen-gdscript/libs` 是所有生成的 `*.pb.gd` 文件都要依赖的
  protobuf 运行时库。需要把这些脚本拷贝到 Godot 项目中一次，让
  `ProtoMessage`、`ProtoUtils`、`ProtoInputFile`、`ProtoOutputBuffer`
  等全局 `class_name` 类能被生成代码直接使用。
- `tools/protoc-gen-gdscript-excel/libs/excel_utils.gd` 是生成
  `*.excel.gd` 时额外需要的运行时辅助库。只使用
  `protoc-gen-gdscript-excel` 还不够，Godot 项目里还必须同时放入上面的
  protobuf `libs`，因为 Excel 包装器会一起调用 `ExcelUtils`、
  `ProtoUtils` 和 `ProtoInputFile`。
- 这些运行时脚本不要求放在某个固定目录下。实际项目里更常见的做法是统一
  放到一个共享目录，例如 `libs` 或 `addons/<name>`，交给 Godot 通过
  `class_name` 完成注册。
- 生成后的 `*.pb.gd` 文件需要尽量保持与源 `.proto` 相同的相对目录结构，
  因为跨 proto 文件的类型引用会生成相对 `preload(...)`。
- 每个 `*.excel.gd` 都需要和对应的 `*.pb.gd` 放在同一个输出目录下。
  生成的 Excel 包装器会从同目录预加载 `./<name>.pb.gd`，而
  `excelc code --gdscript_out=...` 也通常会在这个目录里继续生成
  `tables.gd` 之类的聚合加载脚本。

一个更贴近实际项目的 Godot 布局如下：

```text
res://libs/proto_message.gd
res://libs/proto_utils.gd
res://libs/proto_input_file.gd
res://libs/excel_utils.gd
res://script/gen/excel/example.pb.gd
res://script/gen/excel/example.excel.gd
res://script/gen/excel/tables.gd
```

一个典型的生成流程如下：

```bash
protoc --gdscript_out=./script/gen path/to/example.proto
protoc --gdscript_out=deterministic=true:./script/gen path/to/example.proto
protoc --gdscript_out=./script/gen --gdscript-excel_out=./script/gen path/to/example.proto
excelc code --pb_dir=./excel_proto --pb_package=excel --gdscript_out=./script/gen/excel
```

需要生成稳定的 map 字段序列化结果时，可传 `deterministic=true`。

## 目录说明
| 路径 | 职责 |
| --- | --- |
| [`./addins/goscr`](./addins/goscr) | 服务级 Go 脚本 add-in、脚本实体/组件辅助与生命周期桥接 |
| [`./addins/goscr/dynamic`](./addins/goscr/dynamic) | 动态脚本加载、工程/方案管理与热更新支持 |
| [`./addins/goscr/fwlib`](./addins/goscr/fwlib) | 向脚本运行时导出的 `core`、`framework`、`scaffold` 符号库 |
| [`./addins/propview`](./addins/propview) | 运行时属性视图、属性同步、序列化和复制辅助 |
| [`./tools/excelc`](./tools/excelc) | Excel 结构生成、访问代码生成和数据导出 CLI |
| [`./tools/excelc/excelutils`](./tools/excelc/excelutils) | 供生成代码使用的哈希/索引转换、表加载与查找辅助 |
| [`./tools/propc`](./tools/propc) | 属性同步代码生成器 |
| [`./tools/protoc-gen-go-excel`](./tools/protoc-gen-go-excel) | 面向 Excel 表访问的 Go Protobuf 插件 |
| [`./tools/protoc-gen-go-structure`](./tools/protoc-gen-go-structure) | 面向深拷贝辅助的 Go Protobuf 插件 |
| [`./tools/protoc-gen-go-variant`](./tools/protoc-gen-go-variant) | 面向 GAP variant 集成的 Go Protobuf 插件 |
| [`./tools/protoc-gen-gdscript`](./tools/protoc-gen-gdscript) | 面向消息类型与序列化逻辑的 GDScript Protobuf 插件 |
| [`./tools/protoc-gen-gdscript-excel`](./tools/protoc-gen-gdscript-excel) | 面向 Excel 表包装器的 GDScript Protobuf 插件 |

## 工具链说明
- `tools/excelc` 基于 Cobra/Viper，拆分为 `proto`、`code`、`data` 三个子命令。
- `tools/propc` 读取 Go 声明文件，并在同目录输出对应的 `*.sync.gen.go` 文件。
- `tools/excelc/examples` 提供了 Excel 配表示例工作簿。

## 安装
如果要在业务代码中引入 add-in 包：

```bash
go get git.golaxy.org/scaffold@latest
```

按需安装命令行工具：

```bash
go install git.golaxy.org/scaffold/tools/excelc@latest
go install git.golaxy.org/scaffold/tools/propc@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-excel@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-structure@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-go-variant@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-gdscript@latest
go install git.golaxy.org/scaffold/tools/protoc-gen-gdscript-excel@latest
```

## 相关仓库
- [Golaxy Distributed Service Development Framework Core](https://github.com/pangdogs/core)
- [Golaxy Distributed Service Development Framework](https://github.com/pangdogs/framework)
