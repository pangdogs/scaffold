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

安装整个 Go 模块依赖：

```bash
go get git.golaxy.org/scaffold@latest
```

安装官方 Go Protobuf 插件和本仓库中的全部生成工具：

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

确保 `$GOBIN` 或 `$(go env GOPATH)/bin` 位于 `PATH`。`protoc-gen-<name>` 会被 `protoc` 自动映射为 `--<name>_out`，插件选项通过 `--<name>_opt` 传入。

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

Excel 流程由三个阶段和三个生成工具配合完成：

```text
.xlsx
  │
  ├─ excelc proto ───────────────> .proto
  │                                  │
  │                                  └─ protoc
  │                                      ├─ *.pb.go / *.structure.go / *.excel.go
  │                                      ├─ *.pb.gd / *.excel.gd
  │                                      └─ *.protoset
  │
  ├─ excelc code + *.protoset ───> tables.go / tables.gd
  │
  └─ excelc data + *.protoset ───> *.json / *.bin / *.bin.idx + *.bin.chk_*
```

这里最容易遗漏的是 `*.protoset`：`excelc proto` 只生成 `.proto`，随后必须用 `protoc --descriptor_set_out` 为 `excelc.proto` 和每个工作簿 proto 分别生成同名 `.protoset`。`excelc code` 和 `excelc data` 都从这些 descriptor set 恢复动态消息和自定义 option。

### 推荐目录

下面是一个不绑定具体业务的前后端分离布局：

```text
config/excel/                    # .xlsx 源文件
build/excel/server/proto/        # 服务端 .proto + .protoset
build/excel/client/proto/        # 客户端 .proto + .protoset
server/gen/excel/                # Go protobuf、查询与聚合加载代码
server/res/excel/                # 服务端 JSON / 二进制数据
client/addons/proto/             # GDScript Protobuf 运行时
client/addons/excel/             # GDScript Excel 运行时
client/script/gen/excel/         # *.pb.gd、*.excel.gd、tables.gd
client/excel/                    # 客户端表数据
```

服务端和客户端建议使用不同的 proto 目录，因为 `--targets` 会裁剪字段，索引结构和 GDScript 内部数组配置也可能不同。

### 工作簿规范

[`tools/excelc/examples/ExampleCN.xlsx`](./tools/excelc/examples/ExampleCN.xlsx) 和 [`ExampleEN.xlsx`](./tools/excelc/examples/ExampleEN.xlsx) 提供了完整示例。

#### 工作簿结构

- 可选的 `@types` 页签声明工作簿内可复用的结构和枚举。
- 普通页签声明表。列名首字符不是字母的列会被忽略，通常可用 `#` 开头的列写注释。
- `@types` 的 `Meta` 支持 `separator`、`scope` 和 `pb_field_number`；索引参数只对普通表字段有效。

所有会生成 Protobuf 标识符的名称，包括 `@types` 的类型名、对象字段名、枚举项名以及普通表字段名，都必须符合 `[A-Za-z][A-Za-z0-9_]*`。名称会先校验，再转换为大驼峰格式。普通表中转换后不是以字母开头的列仍作为注释列忽略。

`@types` 的 `Alias` 列用于配置对象字段别名和枚举值别名。别名可以包含中文，但不能包含 ASCII 空格、YAML 指示字符（`-?:,[]{}#&*!|>'"%@`）、反引号、反斜杠或 Unicode 控制字符，以保证别名能安全地作为未加引号的 YAML 对象 key 和 Protobuf option 值使用。

#### 表页行布局

| 行      | 内容                          |
|--------|-----------------------------|
| 第 1 行  | 字段名。                        |
| 第 2 行  | 字段类型。                       |
| 第 3 行  | 字段 `Meta`，也兼容表头名“元数据”或“特性”。 |
| 第 4 行  | 注释。                         |
| 第 5 行起 | 实际数据。                       |

#### 单元格写法

| 类型          | 示例与说明                                                         |
|-------------|---------------------------------------------------------------|
| 标量          | `1`、`3.14`、`true`、`HelloWorld`。                               |
| `bytes`     | Base64 文本，例如 `SGVsbG9Xb3JsZA==`。                              |
| 枚举          | 可写数值、枚举名或别名。                                                  |
| repeated 标量 | 默认用 `,` 分隔，例如 `1,2,3`；可通过 `separator=\|` 改成 `1\|2\|3`。        |
| 对象          | YAML 风格映射，例如 `id: 1, name: Example, tags: [1, 2]`；字段名和别名都可使用。 |
| repeated 对象 | YAML 数组 `[{id: 1}, {id: 2}]`，或配合自定义分隔符书写多个映射。                 |
| map         | YAML 风格映射，例如 `1: Alpha, 2: Beta` 或 `1: {id: 1, name: Alpha}`。 |

对象映射的 `:` 后必须留有空白，例如 `name: Example`。对象字段名不允许包含 `:`；导出数据时，`excelc` 会拒绝任何包含 `:` 的对象 key。检查会递归作用于对象、repeated 对象和 map 中的对象值，但不会扫描普通字符串值或 map key。不含 `:` 的未知对象字段仍会被忽略，以兼容 `targets`/`scope` 对 schema 的裁剪。

#### 字段 Meta

Meta 使用 query-string 格式，例如 `scope=client&sorted_unique_index=1`：

| 参数                    | 说明                                                      |
|-----------------------|---------------------------------------------------------|
| `scope`               | 可重复的目标标签，与 `--targets` 配合裁剪字段；未设置时对所有目标可见。              |
| `separator`           | repeated 或 map 单元格的自定义分隔符，默认 `,`。                       |
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
