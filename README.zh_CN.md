# Scaffold

[English](./README.md) | [简体中文](./README.zh_CN.md)

## 项目简介

`scaffold` 是 [Golaxy Distributed Service Development Framework](https://github.com/pangdogs/framework) 的配套工具与 add-in 仓库，为 Go 服务端、Godot 客户端和 Excel 配表流程提供代码生成、数据导出、运行时集成、属性同步与脚本热更新能力。

这个仓库不是独立的业务框架，主要包含三类组件：

- `tools`：构建期运行的命令行工具和 `protoc` 插件。
- `addins`：挂接到 `git.golaxy.org/framework` 的 Go 运行时扩展。
- `godot`：可拷贝到 Godot 4 项目中的客户端运行时脚本。

## 文档导航

- [组件一览](#组件一览)
- [环境与安装](#环境与安装)
- [Protobuf 代码生成](#protobuf-代码生成)
- [Excel 配表工具链](#excel-配表工具链)
- [属性同步代码生成](#属性同步代码生成)
- [运行时组件](#运行时组件)
- [仓库目录](#仓库目录)

## 组件一览

| 类型        | 组件                                      | 主要职责                                              |
|-----------|-----------------------------------------|---------------------------------------------------|
| Go add-in | `addins/goscr`                          | 基于 Yaegi 加载 Go 脚本工程，支持脚本化实体 / 组件声明、本地或远端源码更新与热重载。 |
| Go add-in | `addins/propview`                       | 托管实体属性的加载、保存、revision 推进以及跨服务或客户端同步。              |
| CLI       | `tools/propc`                           | 扫描带注解的 Go 属性声明并生成 `*.sync.gen.go`。                |
| CLI       | `tools/excelc`                          | 从 `.xlsx` 生成表结构 proto、聚合访问代码以及 JSON / 二进制数据。      |
| protoc 插件 | `tools/protoc-gen-go-structure`         | 为 Go Protobuf 消息生成深拷贝辅助方法。                        |
| protoc 插件 | `tools/protoc-gen-go-variant`           | 让 Go Protobuf 消息实现 Golaxy GAP variant 所需接口。       |
| protoc 插件 | `tools/protoc-gen-go-excel`             | 为 Excel 表消息生成 Go 查询和索引访问方法。                       |
| protoc 插件 | `tools/protoc-gen-gdscript`             | 生成 Godot 侧的 `*.pb.gd` Protobuf 消息代码。              |
| protoc 插件 | `tools/protoc-gen-gdscript-excel`       | 生成 Godot 侧的 `*.excel.gd` 表包装器和索引查询方法。             |
| Godot 运行时 | `tools/protoc-gen-gdscript/godot`       | `*.pb.gd` 依赖的 Protobuf 编解码、文件流、哈希和消息基类。           |
| Godot 运行时 | `tools/protoc-gen-gdscript-excel/godot` | `*.excel.gd` 依赖的索引查询、分块文件和二分查找辅助。                 |
| Godot 运行时 | `godot/rpcli`                           | GAP / GTP 连接、重连、RPC 调用、回调绑定和 variant 传输。          |
| Godot 运行时 | `godot/resty`                           | Resty 风格 HTTP 请求、下载、并发请求和 Server-Sent Events。     |

## 环境与安装

### 前置依赖

- Go `1.25.0` 或与当前 `go.mod` 兼容的版本。
- `protoc`，以及可被 `-I` 引用的 `google/protobuf/descriptor.proto`。
- 生成 Go Protobuf 时还需要官方 `protoc-gen-go`。
- 使用 GDScript 产物时需要 Godot 4，并把对应运行时脚本放入项目。

### 安装 `protoc`

`protoc` 是由官方 [protocolbuffers/protobuf](https://github.com/protocolbuffers/protobuf) 项目发布的 Protocol Buffers 编译器，用于编译 `excelc proto` 生成的 `.proto` schema、生成 Protobuf 绑定代码，以及生成 Excel 工作流所需的 `*.protoset` descriptor set。特别是生成的 `excelc.proto` 会导入 `google/protobuf/descriptor.proto`，因此必须保留编译器发行包中的 `include` 目录。

Windows 推荐使用 Winget 安装；安装后请新开终端再验证：

```powershell
winget install protobuf
protoc --version
```

也可以从[官方 Releases](https://github.com/protocolbuffers/protobuf/releases/latest)下载与操作系统和 CPU 架构匹配的预编译包，例如 Windows x64 使用 `protoc-<version>-win64.zip`。把压缩包解压到自定义目录（下面以 `C:\tools\protobuf` 为例），保留其中的 `bin` 和 `include` 目录，并将 `bin` 加入 `PATH`：

```text
C:\tools\protobuf\
├─ bin\
│  └─ protoc.exe
└─ include\
   └─ google\protobuf\descriptor.proto
```

以下 PowerShell 命令用于配置和验证当前终端中的手动安装：

```powershell
$protocRoot = 'C:\tools\protobuf'
$protocBin = Join-Path $protocRoot 'bin'
$env:Path = "$protocBin;$env:Path"
$env:PROTOBUF_INCLUDE = Join-Path $protocRoot 'include'

protoc --version
Test-Path (Join-Path $env:PROTOBUF_INCLUDE 'google\protobuf\descriptor.proto')
```

`PROTOBUF_INCLUDE`（或 `-I` 参数）必须指向 `include` 根目录，不能指向其 `google/protobuf` 子目录。macOS 和 Debian/Ubuntu 可分别通过 `brew install protobuf` 和 `sudo apt-get install protobuf-compiler` 安装，安装后同样使用 `protoc --version` 验证。

### 安装 Go 模块

如果应用代码会导入本仓库的 add-ins 或生成的运行时辅助包，请将本仓库添加为 Go 模块依赖：

```bash
go get git.golaxy.org/scaffold@latest
```

### 安装代码生成工具

代码生成工具通常通过 Go 安装；如果提交或发布后不希望使用方再安装 Go，也可以把已编译工具归档到项目目录。

#### 通过 Go 安装

开发机推荐直接执行下面的 `go install` 命令。未设置 `GOBIN` 时，Go 通常会把编译出的可执行文件安装到 `$(go env GOPATH)/bin`；请确保该目录以及 `protoc` 所在的 `bin` 目录都位于 `PATH`。

如果需要将 Go 工具安装到自定义目录，在执行命令前设置 `GOBIN`，例如：

```powershell
$generatorBin = 'C:\tools\scaffold\bin'
New-Item -ItemType Directory -Force -Path $generatorBin | Out-Null
$env:GOBIN = $generatorBin
$env:Path = "$generatorBin;$env:Path"
```

然后执行：

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

#### 归档已编译工具

如果提交或发布后不希望使用方再安装 Go，可以把已经编译好的 `.exe` 复制到项目的 `tools` 目录中。推荐按 `garden/tools` 的方式按用途分目录：`protobuf/bin` 放 `protoc.exe` 和 `protoc-gen-*` 插件，`excelc/bin` 放 `excelc.exe`，如果使用 `propc` 则放到独立的 `propc/bin`。`protoc.exe` 和 `include` 仍从官方 `protocolbuffers/protobuf` 发布包获取。

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

在项目根目录执行下面的 PowerShell 命令，即可在当前终端使用这些项目内工具：

```powershell
$projectTools = Join-Path (Get-Location) 'tools'
$protobufBin = Join-Path $projectTools 'protobuf\bin'
$excelcBin = Join-Path $projectTools 'excelc\bin'
$propcBin = Join-Path $projectTools 'propc\bin'
$env:Path = "$protobufBin;$excelcBin;$propcBin;$env:Path"
$env:PROTOBUF_INCLUDE = Join-Path $projectTools 'protobuf\include'
```

确保项目内 `protoc`、`excelc` 和 `propc` 所在的目录都位于 `PATH`。`protoc-gen-<name>` 会被 `protoc` 自动映射为 `--<name>_out`，插件选项通过 `--<name>_opt` 传入。

本仓库的 Protobuf 插件基于 Go `protogen`。即使只生成 GDScript，每个输入 `.proto` 也需要提供合法的 `option go_package`，或者通过 `M<file>=<go-import-path>` 插件参数映射 Go import path。

## Protobuf 代码生成

### 最小 schema

以下示例用于说明普通业务 Protobuf 的生成方式；Excel 生成的 schema 见后面的 [Excel 完整流程](#完整流程)。

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

假设保存为 `proto/example/profile.proto`。

### 生成 Go 代码

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

这条命令分别生成：

| 文件                     | 生成器                       | 内容                        |
|------------------------|---------------------------|---------------------------|
| `profile.pb.go`        | `protoc-gen-go`           | 官方 Go Protobuf 消息。        |
| `profile.structure.go` | `protoc-gen-go-structure` | 消息和字段级深拷贝方法。              |
| `profile.variant.go`   | `protoc-gen-go-variant`   | GAP variant 注册、类型编号及读写接口。 |

### `protoc-gen-go-structure`

该插件必须和 `protoc-gen-go` 对同一批 schema、同一个 Go 包运行。它不生成消息定义，只为文件中的顶层 message 补充：

- `Clone()` 消息深拷贝。
- `Clone<Field>()` 字段级拷贝。
- 对消息、列表、map、`bytes` 和标量字段采用对应的深拷贝策略。

插件没有自定义选项，只使用 `protogen` 通用的 `paths`、`module` 和 `M<file>=<import>` 等参数。

### `protoc-gen-go-variant`

该插件让文件中的每个顶层 Go 消息参与 Golaxy GAP variant 体系，输出 `*.variant.go`，包括：

- 在 `init()` 中注册消息类型。
- `Read`、`Write`、`Size`、`TypeId` 和 `Indirect` 等方法。
- 基于 Protobuf 完整消息名生成稳定的自定义 variant type id。

| 选项              | 默认值     | 说明                                     |
|-----------------|---------|----------------------------------------|
| `deterministic` | `false` | 使用确定性 Protobuf 序列化；需要跨进程稳定字节序或哈希时建议开启。 |

生成代码依赖 `git.golaxy.org/framework/net/gap/variant`。

Go 和 GDScript variant 使用兼容的 32 位 FNV-1a type id。它对相同的 `package.message` 保持稳定，但生成器不会跨全部 schema 检测哈希碰撞；协议规模很大时应在构建或测试阶段校验 type id 唯一性。

### 生成 GDScript 代码

```bash
protoc \
  -I./proto \
  --gdscript_out=./client/script/gen \
  --gdscript_opt=paths=source_relative \
  --gdscript_opt=string_as_string_name=true \
  --gdscript_opt=deterministic=true \
  ./proto/example/profile.proto
```

输出 `profile.pb.gd`。生成文件不是自包含的，还需要把 [`tools/protoc-gen-gdscript/godot`](./tools/protoc-gen-gdscript/godot) 中的全部脚本拷贝到 Godot 项目，例如 `res://addons/proto/`。

### `protoc-gen-gdscript`

该插件为 proto3 的枚举和消息生成：

- GDScript 字段、嵌套消息和枚举。
- 二进制序列化、反序列化和大小计算。
- `reset`、`clone`、`equals`、`hash`、`to_dict` 和 `from_dict`。
- repeated、map、packed/unpacked 数值字段与跨文件 `preload`。

| 选项                      | 默认值     | 说明                                                             |
|-------------------------|---------|----------------------------------------------------------------|
| `string_as_string_name` | `false` | 将 proto `string` 映射为 `StringName`；适合大量重复标识符，不适合任意长文本。          |
| `deterministic`         | `false` | 对 map key 排序后序列化，使相同消息产生稳定字节序。                                 |
| `gap_variant`           | `false` | 让消息继承 `ProtoGAPVariant` 并注册到 `GAPVariants`；同时依赖 `godot/rpcli`。 |
| `class_name`            | `false` | 为生成文件导出顶层 `class_name`；默认通过脚本 `preload` 使用。                    |

跨文件引用使用相对 `preload(...)`，因此应让生成目录保持源 `.proto` 的相对层级。启用 `gap_variant=true` 时，还要把 [`godot/rpcli`](./godot/rpcli) 放入项目，使 `GAPVariants` 可用。

#### Protobuf 支持范围

`protoc-gen-gdscript` 仅面向 `syntax = "proto3"`，不支持 proto2 和 Protobuf Editions。常规标量、枚举、消息、`repeated`、`map` 与 packed/unpacked 数组均可使用，但有以下边界：

| 功能                    | 当前行为与限制                                                                                                                         |
|-----------------------|---------------------------------------------------------------------------------------------------------------------------------|
| Unknown fields        | 可以跳过 varint、fixed32、fixed64 和 length-delimited 未知字段，但不会保存；再次序列化时这些字段会丢失。未知 group 字段无法跳过。                                        |
| `oneof`               | 不生成 case/discriminator，也不执行成员互斥或自动清除；成员会被当成独立普通字段，不应在 GDScript 目标中使用。                                                           |
| `optional` / presence | proto3 标量 `optional` 不生成 `has_*` / `clear_*`；无法区分“未设置”和“显式设置为默认值”。消息字段仍可用 `null` 表示不存在。                                         |
| Services              | 不根据 `service` 声明生成 RPC client/server stub。                                                                                      |
| ProtoJSON             | `to_dict` / `from_dict` 是便捷转换，不是完整 ProtoJSON 实现；未提供 `Any`、`Timestamp`、`Duration`、`FieldMask`、`Struct` 等 well-known types 的特殊映射。 |
| `uint64` / `fixed64`  | wire 编解码保留完整 64 位，但 Godot `int` 是有符号 64 位；大于 `9223372036854775807` 的值表现为负数。业务字段应尽量限制在 int64 正数范围。                               |

## Excel 配表工具链

### 工具关系与产物

Excel 工具链分为结构、代码和数据三个阶段。`excelc proto` 从工作簿生成 Protobuf 结构，`protoc` 及其插件生成各语言的单表代码，`excelc code` 生成所有表的聚合入口，最后由 `excelc data` 导出运行时数据：

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

这里最容易遗漏的是 `*.protoset`：`excelc proto` 只生成 `.proto`，随后必须用 `protoc --descriptor_set_out` 为 `excelc.proto` 和每个工作簿 proto 分别生成同名 `.protoset`。`excelc code` 和 `excelc data` 都从这些 descriptor set 恢复动态消息和自定义 option。

以 `Config.xlsx` 为例，默认会生成 `Config.proto`，其中声明行消息 `ConfigColumns` 和表消息 `ConfigTable`。对应的单表代码文件是 `Config.pb.go`、`Config.structure.go`、`Config.excel.go`（Godot 侧为 `Config.pb.gd`、`Config.excel.gd`）；运行时数据使用表消息名，因而文件名是 `ConfigTable.json`、`ConfigTable.bin` 或 `ConfigTable.bin.idx`。

#### 产物职责

| 产物 | 生成方式 | 作用 |
|------|---------|------|
| `excelc.proto` | `excelc proto` | 所有表共享的声明，包含 Excel custom options、索引结构和分块清单等基础消息。每套目标 schema 生成一次。 |
| `<Workbook>.proto` | `excelc proto` | 单个工作簿的静态结构，包含 `*Columns`、`*Table`、工作簿内对象和枚举，以及表使用的索引字段；它不包含数据行。 |
| `*.protoset` | `protoc --descriptor_set_out` | `.proto` 的 descriptor set。`excelc code` 和 `excelc data` 用它识别消息、字段、scope 和索引 option；仅在构建/打表阶段使用。 |
| `*.pb.go` | `protoc-gen-go` | Go 侧的 Protobuf 消息、枚举和 wire 编解码类型；表数据最终反序列化到这里定义的 `*Table` 和 `*Columns`。 |
| `*.structure.go` | `protoc-gen-go-structure` | 可选的 Go 深拷贝与字段克隆辅助方法，不负责加载或查询表。 |
| `*.excel.go` | `protoc-gen-go-excel` | 单张表的 Go 查询代码，在 `*Table` 上生成唯一/非唯一索引的 `Lookup`、`Get` 和 `LookupBy...` 方法。 |
| `tables.go` | `excelc code --go_out` | Go 聚合入口。`Tables` 为每张表提供一个字段；`LoadJsonFiles` 或 `LoadBinaryFiles` 从同一目录加载全部表并返回该容器。 |
| `*.pb.gd` | `protoc-gen-gdscript` | Godot 侧的 Protobuf 消息、枚举以及序列化/反序列化代码，对应 Go 的 `*.pb.go`。 |
| `*.excel.gd` | `protoc-gen-gdscript-excel` | 单张表的 GDScript 包装器，提供行访问、索引查询以及分块表的同步/异步访问能力。 |
| `tables.gd` | `excelc code --gdscript_out` | Godot 聚合入口。它预加载各表的 `*.pb.gd` / `*.excel.gd`，统一导出消息和枚举，保存每张表的包装器，并由 `load_data()` 加载普通或分块二进制。通常注册为 autoload。 |
| `*Table.json` | `excelc data --json_out` | 可读的 Protobuf JSON 表数据，包含行和索引；主要供 Go 的 `LoadJsonFiles`、检查和热加载使用。 |
| `*Table.bin` | `excelc data --binary_out` | 完整的 Protobuf 二进制表消息，行和索引都在一个文件内，可由 Go 或 Godot 加载。 |
| `*Table.bin.idx` | `excelc data --binary_out --binary_chunked` | 分块模式的入口文件，保存索引、chunk manifest 等信息，不保存 `Rows`。 |
| `*Table.bin.chk_N` | 同上 | 分块模式的数据文件，只保存对应范围的 `Rows`；Godot 包装器根据查询或行访问按需加载。 |

`tables.go` 和 `tables.gd` 都不保存实际表数据，而是把“加载全部表”和“访问每张表”集中到一个入口。以生成的 `ConfigTable` 为例，运行关系如下：

```text
Go：
ConfigTable.json / ConfigTable.bin
  └─ LoadJsonFiles / LoadBinaryFiles
       └─ tables *Tables
            └─ tables.ConfigTable.Lookup(...)

Godot 普通二进制：
Excel.load_data()
  └─ 读取 ConfigTable.bin（行与索引）
       └─ 创建 Excel.ConfigTable
            └─ Excel.ConfigTable.lookup(...)

Godot 分块二进制：
Excel.load_data()
  └─ 只读取 ConfigTable.bin.idx（索引与 chunk manifest）
       └─ 创建 Excel.ConfigTable（内部为 ConfigChunkedTable）
            └─ Excel.ConfigTable.lookup(...) / await Excel.ConfigTable.lookup_async(...)
                 └─ 根据索引得到 row offset
                      └─ 按需读取并缓存对应的 ConfigTable.bin.chk_N
```

上面的 `Excel` 是将 `tables.gd` 注册为 autoload 后的实例名；不使用 autoload 时也可以自行创建 `Tables` 实例。分块表的同步和异步查询都会按需加载 chunk。同步方法在调用线程完成首次加载，可在主线程或非主线程调用；异步方法只允许在主线程调用，在线程可用时会把加载提交到工作线程并在等待期间让出执行权。无线程平台会在主线程内联完成异步加载，因此首次访问仍会同步阻塞。调用 `rows()` / `rows_async()` 时则需要加载所有 chunk。

`.xlsx`、`.proto` 和 `*.protoset` 属于配置或构建输入，通常不随程序发布。Go 项目编译生成的 `*.go`，并按所选加载方式部署 JSON 或二进制数据；Godot 项目需要生成的 `*.gd`、两套 Godot 运行时脚本，以及二进制表数据。

### 推荐目录

下面是一个不绑定具体业务的前后端分离布局：

```text
config/excel/                    # 配表源文件：*.xlsx
build/excel/server/proto/        # 服务端构建中间产物：*.proto + *.protoset
build/excel/client/proto/        # 客户端构建中间产物：*.proto + *.protoset
server/gen/excel/                # *.pb.go、*.structure.go、*.excel.go、tables.go
server/res/excel/                # 运行时加载的 *Table.json 或 *Table.bin
client/addons/proto/             # tools/protoc-gen-gdscript/godot 运行时
client/addons/excel/             # tools/protoc-gen-gdscript-excel/godot 运行时
client/script/gen/excel/         # *.pb.gd、*.excel.gd、tables.gd
client/excel/                    # *Table.bin，或 *Table.bin.idx + *Table.bin.chk_*
```

服务端和客户端建议使用不同的 proto 目录，因为 `--targets` 会裁剪字段，索引结构和 GDScript 内部数组配置也可能不同。

### 工作簿规范

[`tools/excelc/examples/ExampleCN.xlsx`](./tools/excelc/examples/ExampleCN.xlsx) 和 [`ExampleEN.xlsx`](./tools/excelc/examples/ExampleEN.xlsx) 提供了完整示例。

两个工作簿使用相同结构，集中展示对象与枚举声明、字段别名、基础类型、列表、对象列表、map、字符串转义、`scope=c` / `scope=s`、单列与复合 `unique_index` / `index`，以及多数据分页和备注页。数据行只填写当前示例需要的字段，未填写字段保持 Protobuf 默认值。

#### 工作簿组成

- 一个工作簿对应一张逻辑表。
- 可选的 `@types` 页签用于声明工作簿内复用的对象和枚举。
- 中文或英文字母开头的普通页签参与数据导出，多个分页按页签顺序合并；不参与数据导出的备注页可用 `#` 开头。

#### `@types` 类型页

`@types` 第 1 行是列名，第 2 行起每行声明一个对象字段或枚举项。同一类型可以连续使用多行；“枚举值”为空时按对象字段解析，有值时按枚举项解析，同一类型不能混用两种形式。

| 列                  | 说明                                                       |
|--------------------|----------------------------------------------------------|
| `对象类型` / `类型`     | 类型名。同一对象的字段或同一枚举的枚举项使用相同类型名。                           |
| `字段名`              | 对象字段名或枚举项名。                                              |
| `字段类型`             | 对象字段的类型；支持内置类型、已声明类型以及 `Type[]` 数组。枚举项不使用此列。           |
| `枚举值`              | 留空表示对象字段；填写非负整数表示枚举项。proto3 枚举的第一个枚举项应为 `0`。            |
| `别名`               | 对象字段或枚举项在数据单元格中的可选别名，可使用中文。                            |
| `默认值`              | 预留列，当前不参与 schema 生成或数据导出。                                  |
| `元数据` / `特性` / `Meta` | 支持 `separator`、`scope` 和 `pb_field_number`；索引参数只用于数据分页字段。 |
| `注释`               | 写入生成的 Protobuf 声明。                                          |

所有会生成 Protobuf 标识符的名称，包括类型名、对象字段名、枚举项名和数据分页字段名，都必须符合 `[A-Za-z][A-Za-z0-9_]*`；名称校验后会转换为大驼峰格式。别名可以包含中文，但不能包含 ASCII 空格、YAML 指示字符（`-?:,[]{}#&*!|>'"%@`）、反引号、反斜杠或 Unicode 控制字符。

#### 数据分页

第一个数据分页定义字段、类型、`Meta`、注释和索引。后续分页只使用第 1 行字段名绑定数据，第 2～4 行内容会被忽略，但仍需保留这三行占位，使数据从第 5 行开始：

| 行      | 内容         |
|--------|------------|
| 第 1 行  | 字段名。       |
| 第 2 行  | 字段类型。      |
| 第 3 行  | 字段 `Meta`。  |
| 第 4 行  | 注释。        |
| 第 5 行起 | 实际数据。      |

声明单元格会裁剪首尾空白，并将 CRLF、CR 统一为 LF；注释中的控制字符会转换为 `\n`、`\t` 等可见转义形式。

- 第 1 行中第一个空字段名单元格是有效列的结束标记。该空列及其右侧所有列都会被忽略，即使其他行仍有内容。
- 在结束标记之前，字段名转换后首字符不是字母的列也会被忽略。通常可用 `#` 开头的列保存行内注释，这类列不会截断其右侧的有效列。
- 第 5 行起，如果所有已识别字段列都为空，该行会被跳过且不占用 row offset。空行不会结束分页，其后的非空行仍会继续导出；只在注释列或有效列右侧填写内容的行仍视为空行。
- 只要任一已识别字段列非空，该行就会导出，未填写的字段保持 Protobuf 默认值。空行判断不受 `--targets` 裁剪影响，因此不同目标的 row offset 可以保持一致。
- 后续分页按字段名绑定，列顺序可以不同，也可以省略不需要填写的字段；缺少的字段按空单元格处理并保持 Protobuf 默认值。后续分页不能出现首个分页未定义的字段，也不能重复字段名。
- 分页合并后共用连续的 row offset 和索引。唯一索引会检查跨分页重复值，非唯一索引结果保持分页及原始行顺序。

#### 单元格写法

| 类型          | 示例与说明                                                         |
|-------------|---------------------------------------------------------------|
| 标量          | `1`、`3.14`、`true`、`HelloWorld`。                                      |
| `bytes`       | Base64 文本，例如 `SGVsbG9Xb3JsZA==`。                                     |
| 枚举          | 可写数值、枚举名或别名。                                                       |
| repeated 标量 | 默认用 `,` 分隔，例如 `1,2,3`；也支持 YAML 数组，例如 `[1, 2, 3]`。               |
| 对象          | YAML 风格映射，例如 `id: 1, name: Example, tags: [1, 2]`；字段名和别名都可使用。    |
| repeated 对象 | 可使用 YAML 数组 `[{ID: 1}, {ID: 2}]`，也可用字段的 `separator` 分隔对象片段，例如 `A: 1, B: Hello \| A: 2, B: World`。 |
| map           | 只能使用 YAML 风格映射，例如 `{1: Alpha, 2: Beta}` 或 `{1: {id: 1, name: Alpha}}`；`separator` 不参与 map 解析。 |

对象映射中，字段名与值之间要写成 `字段名: 值`，冒号后必须有空格：

```yaml
name: Example  # 正确
name:Example   # 错误，excelc 会提示在冒号后添加空格
```

该检查适用于对象、repeated 对象以及 map 中作为 value 的对象；普通字符串和 map key 不受影响。无法在当前 schema 中找到的对象字段仍会被忽略，以便同一份数据用于经过 `targets`/`scope` 裁剪的不同 schema。

同一个 YAML mapping 内不允许出现重复 key；检查会递归应用于对象、repeated 对象和 map。例如 `{1: Alpha, 1: Beta}` 会报告 key `1` 的首次出现位置和重复出现位置，不会静默使用其中一个值。错误中的 line 和 column 是当前 Excel 单元格内 YAML 文本的位置，不是工作表的行列坐标；工作表、数据行和字段名会由外层错误信息给出。

#### 值解析规则

- 字段解析开始时会裁剪单元格首尾空白；裁剪后为空的值会被忽略。需要保留首尾空白时，必须使用单引号或双引号，例如 `"  Hello  "`。
- Excel 单元格中直接输入的换行会原样保存到 Protobuf，Go 和 Godot 读取后就是换行；不同系统的换行格式会统一为 LF，单元格开头和结尾的换行会作为首尾空白被裁剪。数据分页中类型为 `string` 的列不会把未加引号的 `\n` 当作换行：`第一行\n第二行` 会原样保留，`"第一行\n第二行"` 中的 `\n` 会按 YAML 双引号转义为换行，`'第一行\n第二行'` 中的 `\n` 仍会原样保留。
- 每个 repeated 字段独立选择解析方式。YAML 根节点是 sequence 时按标准 YAML 数组处理，支持 `[1, 2, 3]` 和块式 `- item`；否则使用该字段自己的 `separator`，默认是 `,`。
- separator 只在引号之外且不处于 `{}`、`[]` 内部时切分。repeated 对象的每个片段会作为独立 YAML mapping 解析，片段外层的 `{}` 可以省略。
- 默认分隔符 `,` 也用于分隔对象字段，因此不适合省略 `{}` 的多字段 repeated 对象。例如 `A: 1, B: Hello, A: 2, B: World` 存在歧义；此时应将对象列表的 separator 改为 `|` 等其他符号，或使用带 `{}` 的对象/标准 YAML sequence。
- 引号可以保护嵌套 repeated 字段的分隔符，使其不参与外层切分；包含对象解析完成后，嵌套字段仍会使用自己的 separator。例如外层和 `D` 都使用 `|` 时，`A: 1, B: HelloWorld, D: "1|2|3" | A: 2, B: HAHAHAHAHA, D: "4|5|6"` 会得到两个对象，两个 `D` 分别包含三项。
- 标准 YAML sequence 中的元素不会再次按 separator 切分。对于 repeated string，`D: ["1|2|3"]` 表示只包含一个字符串；`D: "1|2|3"` 则进入 separator 模式并得到三项。
- 单引号和双引号都属于 YAML 语法，解析后不会写入字符串值；引号内的转义按 YAML 规则处理。任一列表元素解析失败时，整个字段解析失败，不会写入部分结果。

#### 字段 Meta

Meta 使用 query-string 格式，例如 `scope=client&sorted_unique_index=1`：

| 参数                    | 说明                                                      |
|-----------------------|---------------------------------------------------------|
| `scope`               | 可重复的目标标签，与 `--targets` 配合裁剪字段；未设置时对所有目标可见。              |
| `separator`           | repeated 字段单元格的自定义分隔符，默认 `,`，支持多字符；不能为空，且首尾或内部均不能包含空白、控制字符、`: ' " \\ { } [ ]`。map 字段始终按 YAML mapping 解析。 |
| `pb_field_number`     | 覆盖 Protobuf field number；必须为合法正数、不能位于保留区间，并且在同一消息内不能重复。 |
| `unique_index`        | 唯一索引逻辑分组，物理结构由 `--pb_unique_index_as` 决定。               |
| `hash_unique_index`   | 强制使用哈希唯一索引。                                             |
| `sorted_unique_index` | 强制使用有序唯一索引。                                             |
| `index`               | 允许同键多行的索引逻辑分组，物理结构由 `--pb_index_as` 决定。                 |
| `hash_index`          | 强制使用哈希非唯一索引。                                            |
| `sorted_index`        | 强制使用有序非唯一索引。                                            |

### 索引模型

| 索引                    | 主要结构                               | 查询特征                   | 适用场景                     |
|-----------------------|------------------------------------|------------------------|--------------------------|
| `hash_unique_index`   | `hash -> row offset`，另带哈希冲突 bucket | 平均常数时间，唯一命中一行          | 内存充足、查询频繁的服务端。           |
| `sorted_unique_index` | `Values + Offsets`                 | 对 `Values` 二分查找        | 希望减少 map 对象的客户端。         |
| `hash_index`          | `hash -> Offsets` bucket           | 平均常数时间，同键返回多行          | 查询优先且可接受 bucket/list 开销。 |
| `sorted_index`        | `Values + Starts + Offsets`        | 二分 key，再读取连续 offset 区间 | 内存敏感的终端设备。               |

索引配置规则：

- 单列索引只在一个字段配置 tag；复合索引让多个字段复用同一个 tag。
- 同一字段可以参加多个索引，例如 `hash_unique_index=1&hash_unique_index=2`。
- 四种物理索引分别使用独立的 tag 分组，可以复用数字；同一个字段不能把同一个 tag 同时配置给多种索引类型。
- 唯一查询返回一行；非唯一查询返回按原始 row offset 排列的多行。
- 查询结果引用表消息 `Rows` 中的对象，不会克隆行。
- `unique_index` 默认使用 `hash_unique_index`，`index` 默认使用 `sorted_index`，可由命令行统一覆盖。

### 完整流程

下面的命令使用 POSIX shell、中性目录和表名，可直接改成项目路径。示例把服务端索引生成为 hash，把客户端索引生成为 sorted；目标标签 `server` / `client` 只是约定，可以换成任意名称。Windows 批处理或 PowerShell 使用等价的续行和逐文件循环即可。

#### 1. 生成服务端和客户端 schema

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

如果需要处理目录中的全部工作簿，把 `--excel_files` 换成 `--excel_dir=./config/excel`。

#### 2. 生成服务端代码和 descriptor set

以下是 POSIX shell 示例。`PROTOBUF_INCLUDE` 应指向包含 `google/protobuf/descriptor.proto` 的 include 根目录。

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

`"${proto}set"` 会把 `Config.proto` 写成 `Config.protoset`，把依赖定义 `excelc.proto` 写成 `excelc.protoset`。不要把所有 schema 合并成一个任意名称的 descriptor set，`excelc` 会按工作簿名查找文件。

#### 3. 生成客户端代码和 descriptor set

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

`string_as_string_name` 必须在 `protoc-gen-gdscript` 和 `protoc-gen-gdscript-excel` 两边保持一致。`*.excel.gd`、对应的 `*.pb.gd` 和聚合脚本 `tables.gd` 应位于同一输出层级。

#### 4. 导出数据

服务端可以导出便于检查和热加载的 JSON：

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

客户端可以导出按需加载的分块二进制：

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

分块模式生成 `*.bin.idx` 和 `*.bin.chk_*`。索引及 chunk manifest 位于 `.idx`，行数据位于 chunk 文件；查询时只加载命中行所在的 chunk。`--binary_chunk_size` 是最大行数而不是字节数，应根据单行大小和访问模式调整。

#### 5. 自动化建议

- 把“schema / 代码编译”和“数据导出”拆成两个脚本。只有表头、类型、Meta、scope 或索引变化时才需要重新生成 schema 和代码；普通数据行变化只运行 `excelc data`。
- 重新生成前清理旧 `.proto`、`.protoset` 和生成代码，避免已删除工作簿的残留 descriptor 被 `excelc code` 扫描。
- 在 Godot 项目内生成代码时保留仍有对应脚本的 `.uid`，只清理孤立 `.uid`，避免资源 UID 无意义变化。
- 每一步检查退出码并立即停止，避免用旧 descriptor set 继续导出新数据。

### `excelc`

`excelc` 有三个子命令。可随时运行 `excelc <command> --help` 查看完整参数。

#### `excelc proto`

从工作簿生成 `excelc.proto` 和每个工作簿对应的表结构 `.proto`。

| 参数                       | 说明                                                                  |
|--------------------------|---------------------------------------------------------------------|
| `--excel_files`          | 显式输入文件列表，优先于 `--excel_dir`。                                         |
| `--excel_dir`            | 扫描目录中的 Excel 文件。                                                    |
| `--pb_out`               | `.proto` 输出目录。                                                      |
| `--pb_package`           | proto package，默认 `excel`。                                           |
| `--pb_options`           | 写入生成 proto 的 file options，例如 `go_package`。                          |
| `--pb_imports`           | 额外写入生成 proto 的 import。                                              |
| `--pb_custom_options`    | Excel 自定义 option 的编号基数，默认 `10000`；同一套产物必须保持一致。                      |
| `--targets`              | 输出目标标签，并与字段 `scope` 一起裁剪 schema。                                    |
| `--pb_unique_index_as`   | `unique_index` 的默认物理结构：`hash_unique_index` 或 `sorted_unique_index`。 |
| `--pb_index_as`          | `index` 的默认物理结构：`hash_index` 或 `sorted_index`。                      |
| `--gdscript_index_array` | GDScript 索引整数向量使用 `packed_int64` 或 `array`，默认 `packed_int64`。       |

#### `excelc code`

读取 `--pb_dir` 下的 `*.protoset` 并生成聚合加载入口：

- `--go_out` 生成 `tables.go`，包含 `Tables`、`LoadBinaryFiles` 和 `LoadJsonFiles`。
- `--gdscript_out` 生成 `tables.gd`，导出表、消息和枚举，并加载普通或分块二进制。
- `--gdscript_class_name` 控制聚合脚本的 `class_name`，默认 `Tables`；传空值可禁用。
- `--gdscript_default_data_dir` 默认 `res://excel/`。
- `--gdscript_autoload` 默认 `true`，会在 `_ready()` 自动调用 `load_data()`；需要自行安排启动顺序或线程时可设为 `false`。

#### `excelc data`

读取工作簿和匹配的 `*.protoset`，构造动态表消息并导出数据：

| 参数                                   | 说明                                |
|--------------------------------------|-----------------------------------|
| `--excel_files` / `--excel_dir`      | 输入工作簿。                            |
| `--pb_dir` / `--pb_package`          | descriptor set 目录和 proto package。 |
| `--targets`                          | 必须与生成该目录 schema 时使用的目标一致。         |
| `--json_out`                         | 导出 `*.json`。                      |
| `--json_multiline` / `--json_indent` | 控制 JSON 可读格式。                     |
| `--binary_out`                       | 导出 `*.bin`。                       |
| `--binary_chunked`                   | 改为 `.bin.idx + .bin.chk_*` 分块格式。  |
| `--binary_chunk_size`                | 每个 chunk 最大行数，默认 `10000`。         |

### `protoc-gen-go-excel`

该插件只面向 `excelc proto` 生成的 schema，输出 `*.excel.go`。它读取表和索引 custom options，为表消息补充：

- 唯一索引：`LookupBy...` 返回 `(*Row, bool)`，`GetBy...` 未命中时 panic；第一个唯一索引还会生成简写 `Lookup` / `Get`。
- 非唯一索引：`LookupBy...` 返回 `[]*Row`，未命中返回 `nil`（可按空切片使用），不生成 `Get`。
- 单列、复合列、哈希冲突校验以及 hash/sorted 两种索引结构。

插件没有自定义选项。生成代码依赖 [`tools/excelc/excelutils`](./tools/excelc/excelutils)，业务 Go 模块需要依赖本仓库。

### `protoc-gen-gdscript-excel`

该插件只面向 `excelc proto` 生成的 schema，输出 `*.excel.gd`，包含普通表和分块表包装器：

- `rows`、`row_count`、`row_at` 以及对应异步方法。
- 唯一索引 `lookup_by_<index_type>_<fields>` 返回 row 或 `null`。
- 非唯一索引使用相同的 `lookup_by_...` 命名，返回 `Array[Row]`。
- 第一个唯一索引会生成简写 `lookup` / `lookup_async`。
- 分块包装器的异步查询会按需加载目标 chunk。
- 所有异步方法只允许在主线程调用；非主线程调用会记录错误并返回空结果，后台线程应使用同步方法。

| 选项                      | 默认值     | 说明                                      |
|-------------------------|---------|-----------------------------------------|
| `string_as_string_name` | `false` | 必须与同一次生成使用的 `protoc-gen-gdscript` 设置一致。 |

生成代码同时依赖 [`tools/protoc-gen-gdscript/godot`](./tools/protoc-gen-gdscript/godot) 和 [`tools/protoc-gen-gdscript-excel/godot`](./tools/protoc-gen-gdscript-excel/godot)。查询返回的是 `Rows` 中的对象引用，不会克隆。

### GDScript 索引数组

`--gdscript_index_array=packed_int64` 会把索引内部的 `Values`、`Starts` 和 `Offsets` 映射为 `PackedInt64Array`，减少 Variant 容器开销并改善连续访问；`array` 用于兼容依赖 `Array[int]` 的旧运行时。

这个选项：

- 只作用于 Excel 内部索引消息，不影响普通 repeated 字段。
- 不改变 Protobuf wire 格式。
- 不影响 Go 生成代码和服务端数据结构。
- 修改后需要重新运行 `excelc proto`，并重新生成 `*.pb.gd` 和 `*.excel.gd`。

## 属性同步代码生成

### `propc`

`propc` 扫描一个 Go 声明文件中紧邻类型和方法的 `+prop-sync-gen:` 注解，生成相邻的 `*.sync.gen.go`。典型声明如下：

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

执行：

```bash
go generate ./...
```

也可以直接指定文件：

```bash
propc --decl_file=profile_prop.go
```

生成的 `profile_prop.sync.gen.go` 会创建 `ProfilePropSync`，包装 `Load`、`Save`、`Managed` 和被标记的方法。同步方法先调用原始实现，再推进 revision 并通过 `propview` 广播操作。

注意事项：

- 注解必须紧邻目标类型或方法的上一行。
- 只有 `sync=true` 的类型会生成包装器。
- 只有该类型的指针 receiver 方法可以标记为同步操作。
- 属性底层状态通常是实现了 GAP `variant.Value` 的消息，可配合 `protoc-gen-go-variant` 生成。
- `//go:generate propc` 会使用 Go 自动提供的 `GOFILE`；手动运行时使用 `--decl_file`。

## 运行时组件

### Go add-ins

#### `addins/propview`

`propview` 提供受管属性表、序列化、revision、跨服务加载 / 保存与增量同步。它通常与 `propc` 生成的 `*Sync` 类型一起使用，并通过 `propview.AddIn` 安装到 Golaxy runtime。

#### `addins/goscr`

`goscr` 是基于 Yaegi 的服务级脚本 add-in，可配置一个或多个本地或远端脚本工程，并把脚本实体 / 组件接入 Golaxy 生命周期。`addins/goscr/dynamic` 负责工程、方案和热更新管理，`addins/goscr/fwlib` 提供导出到脚本环境的符号库。

### Godot 运行时目录

| 目录                                      | 何时需要                                              |
|-----------------------------------------|---------------------------------------------------|
| `tools/protoc-gen-gdscript/godot`       | 任何生成的 `*.pb.gd`。                                  |
| `tools/protoc-gen-gdscript-excel/godot` | 任何生成的 `*.excel.gd`；同时仍需要上一项。                      |
| `godot/rpcli`                           | Godot 连接 Golaxy GAP / GTP，或启用 `gap_variant=true`。 |
| `godot/resty`                           | Godot 发起普通 HTTP、下载或 SSE 请求。                       |

这些目录没有固定安装路径，常见布局如下：

```text
res://addons/proto/
res://addons/excel/
res://addons/rpcli/
res://addons/resty/
res://script/gen/proto/
res://script/gen/excel/
res://excel/
```

生成的 Excel 聚合脚本可注册为 autoload：

```ini
[autoload]

Excel="*res://script/gen/excel/tables.gd"
```

如果生成时设置了 `--gdscript_autoload=false`，由启动流程显式调用：

```gdscript
if !Excel.load_data("res://excel/"):
	push_error("failed to load excel data")
```

RPC 客户端可注册为 autoload 后连接服务：

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

HTTP 客户端可注册为 autoload 后创建独立请求快照：

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

`Resty` 还支持 JSON / 表单 / 原始请求体、路径参数、输出文件、并发请求句柄和 `Resty.sse(url)`。

## 仓库目录

| 路径                                                                     | 职责                        |
|------------------------------------------------------------------------|---------------------------|
| [`addins/goscr`](./addins/goscr)                                       | Go 脚本 add-in、动态工程与热更新。    |
| [`addins/propview`](./addins/propview)                                 | 受管属性和跨端同步。                |
| [`tools/excelc`](./tools/excelc)                                       | Excel schema、代码和数据生成 CLI。 |
| [`tools/excelc/examples`](./tools/excelc/examples)                     | Excel 工作簿示例。              |
| [`tools/excelc/excelutils`](./tools/excelc/excelutils)                 | Go 表加载、索引、哈希和比较辅助。        |
| [`tools/propc`](./tools/propc)                                         | 属性同步代码生成器。                |
| [`tools/protoc-gen-go-structure`](./tools/protoc-gen-go-structure)     | Go Protobuf 深拷贝插件。        |
| [`tools/protoc-gen-go-variant`](./tools/protoc-gen-go-variant)         | Go GAP variant 插件。        |
| [`tools/protoc-gen-go-excel`](./tools/protoc-gen-go-excel)             | Go Excel 查询插件。            |
| [`tools/protoc-gen-gdscript`](./tools/protoc-gen-gdscript)             | GDScript Protobuf 插件与运行时。 |
| [`tools/protoc-gen-gdscript-excel`](./tools/protoc-gen-gdscript-excel) | GDScript Excel 插件与运行时。    |
| [`godot/rpcli`](./godot/rpcli)                                         | Godot GAP / GTP RPC 客户端。  |
| [`godot/resty`](./godot/resty)                                         | Godot HTTP / SSE 客户端。     |

## 相关仓库

- [Golaxy Distributed Service Development Framework Core](https://github.com/pangdogs/core)
- [Golaxy Distributed Service Development Framework](https://github.com/pangdogs/framework)

## 许可证

本项目采用 [GNU Lesser General Public License v2.1](./LICENSE)。
