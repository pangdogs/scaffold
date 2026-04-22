# Scaffold
[English](./README.md) | [简体中文](./README.zh_CN.md)

## 简介
`scaffold` 是
[**Golaxy Distributed Service Development Framework**](https://github.com/pangdogs/framework)
配套的工具与 add-in 仓库，聚焦主运行时之外的项目脚手架能力，包括：

- Excel 配表编译与数据导出
- Protobuf / GDScript 代码生成
- Go 脚本热更新支持
- 实体属性同步相关工具

仓库主要分为两层：

- `addins`：挂接到 `git.golaxy.org/framework` 的服务级或运行时扩展
- `tools`：构建期使用的命令行生成器和 Protobuf 插件

## 提供的能力
- `addins/goscr`：基于 Yaegi 的服务级 Go 脚本 add-in，支持脚本化实体/组件声明、脚本工程加载，以及本地或远端脚本热更新。
- `addins/propview`：运行时属性视图 add-in，负责实体属性的托管、加载、保存和跨服务或客户端同步。
- `tools/propc`：属性同步代码生成器，扫描带注解的 Go 声明并生成基于 `propview` 的同步包装代码。
- `tools/excelc`：把 `.xlsx` 表格转换为 Proto 结构、访问代码，以及二进制 / JSON 数据文件的 Excel 工具链。
- `tools/protoc-gen-go-excel`：为生成的 Go Protobuf 代码补充表查询与索引访问函数。
- `tools/protoc-gen-go-structure`：为 Protobuf 消息补充深拷贝辅助函数，覆盖消息、切片、映射、字节数组和嵌套消息。
- `tools/protoc-gen-go-variant`：让生成的 Protobuf 消息可以作为 Golaxy GAP / RPC 栈中的 variant 值使用。
- `tools/protoc-gen-gdscript`：生成 Godot 侧可用的 GDScript Protobuf 消息类型与序列化逻辑。
- `tools/protoc-gen-gdscript-excel`：为 Excel 生成的表结构补充 GDScript 表包装器和索引查询函数。

## 典型工作流
### Protobuf Schema 流水线
1. 编写共享的 `.proto` schema，用于 RPC、持久化、快照、配置或其他跨进程数据契约。
2. 使用 `protoc` 按需生成不同运行时的绑定代码，例如 Go 服务端代码、结构拷贝辅助、GAP variant 集成，或 GDScript 客户端代码。
3. 将生成结果分别输出到不同运行时目录，例如 `./server/src/gen` 和 `./client/script/gen`。

### Excel 配表流水线
1. 编写 `.xlsx` 表格文件。
2. 使用 `excelc proto` 生成面向表结构的 `.proto` 文件。
3. 对这些生成出来的 schema 继续执行常规 protobuf 生成流程。
4. 使用 `excelc data` 导出二进制和 / 或 JSON 数据文件。

Excel 流水线是建立在 protobuf 之上的。生成出来的表结构 schema 仍然是普通的 protobuf 消息，因此除了表查询之外，也可以用于传输、序列化、存储、快照或其他工具链场景。

### 属性同步
1. 在 Go 中定义属性状态类型和需要同步的方法。
2. 按 `tools/propc` 的约定添加注解并生成 `*.sync.gen.go`。
3. 通过 `addins/propview` 声明属性，让其具备加载、保存和按 revision 复制的能力。

### 脚本热更新
1. 在服务端安装 `addins/goscr`。
2. 通过 `goscr.With.Projects(...)` 指定一个或多个本地或远端脚本工程。
3. 使用 `goscr.BuildEntityPT(...)` 结合脚本元信息声明实体原型。
4. 由 add-in 自动监听本地文件变更或远端源码变更并重新加载脚本方案。

## Godot 运行时库
### Protobuf 运行时
- `tools/protoc-gen-gdscript/libs` 是所有生成的 `*.pb.gd` 文件都依赖的 protobuf 运行时库。需要把这个目录拷贝到 Godot 项目中一次，让 `ProtoMessage`、`ProtoUtils`、`ProtoInputFile`、`ProtoOutputBuffer` 等全局类可被生成代码直接使用。

### Excel 表格运行时
- `tools/protoc-gen-gdscript-excel/libs` 是生成 `*.excel.gd` 包装器时额外需要的运行时目录。
- Excel 表格包装器仍然依赖上面的 protobuf 运行时，因此只要使用 `protoc-gen-gdscript-excel`，这两套运行时目录都需要同时放进 Godot 项目。

### 布局约束
- 这些运行时脚本不要求放在固定目录。实际项目里更常见的做法是统一放到 `libs` 或 `addons/<name>` 一类目录下，通过 `class_name` 注册给 Godot。
- 生成后的 `*.pb.gd` 文件应尽量保持与源 `.proto` 文件相同的相对目录结构，因为跨文件 protobuf 引用会生成相对 `preload(...)` 调用。
- 普通业务 protobuf 输出与 Excel 派生 protobuf 输出通常应放在不同的根目录下维护。通信 / 存储协议和表格协议一般不是同一套产物，不建议混在一个输出目录里。
- 每个 `*.excel.gd` 文件都应和对应的 `*.pb.gd` 放在同一输出目录下。生成的 Excel 包装器会从同目录预加载 `./<name>.pb.gd`，而 `excelc code --gdscript_out=...` 也通常会在该目录中生成 `tables.gd` 之类的聚合加载脚本。

## Godot 项目布局示例
### Protobuf 流水线
Godot 项目中，普通 protobuf 产物一种常见布局如下：

```text
res://addons/proto/          # 从 tools/protoc-gen-gdscript/libs 拷贝的运行时文件
res://script/gen/proto/      # 普通 protobuf 生成的客户端消息
res://script/gen/proto/login.pb.gd
```

### Excel 配表流水线
Godot 项目中，Excel 配表产物一种常见布局如下。这里仅列出 Excel 配表这一侧的目录：

```text
res://addons/proto/          # 从 tools/protoc-gen-gdscript/libs 拷贝的运行时文件
res://addons/excel/          # 从 tools/protoc-gen-gdscript-excel/libs 拷贝的运行时文件
res://script/gen/excel/      # Excel protobuf 与包装器输出
res://script/gen/excel/excelc.pb.gd
res://script/gen/excel/example.pb.gd
res://script/gen/excel/example.excel.gd
res://script/gen/excel/tables.gd
res://res/excel/             # 导出的 Excel 数据文件
res://res/excel/ExampleTable.bin.idx
res://res/excel/ExampleTable.bin.chk_0
```

对于前后端分离的 Excel 项目，一种比较顺手的目录层级可以是：

```text
./config/excel/              # Excel 源文件
./excelc/server/proto/       # 服务端使用的 Excel proto
./excelc/client/proto/       # 客户端使用的 Excel proto
./server/src/gen/            # 服务端生成代码
./server/res/excel/          # 服务端导出的表数据
./client/script/gen/         # 客户端生成代码
./client/res/excel/          # 客户端导出的表数据
```

按这种结构，编译 Excel proto、生成访问代码、导出表数据的流程可以写成：

```bash
# 1. 导出服务端 proto，唯一索引用 hash_unique_index。
excelc proto \
  --excel_files=./config/excel/Consts.xlsx \
  --excel_dir=./config/excel \
  --pb_out=./excelc/server/proto \
  --pb_package=excel \
  --pb_options=[go_package=./excel] \
  --pb_imports=Consts.proto \
  --pb_unique_index_as=hash_unique_index \
  --targets=s

# 2. 导出客户端 proto，唯一索引用 sorted_unique_index。
excelc proto \
  --excel_files=./config/excel/Consts.xlsx \
  --excel_dir=./config/excel \
  --pb_out=./excelc/client/proto \
  --pb_package=excel \
  --pb_options=[go_package=./excel] \
  --pb_imports=Consts.proto \
  --pb_unique_index_as=sorted_unique_index \
  --targets=c

# 3. 生成服务端 Go protobuf 与 Excel 查询代码。
protoc -I./excelc/server/proto -I./protobuf/include \
  --include_imports \
  --retain_options \
  --go_out=./server/src/gen \
  --go-structure_out=./server/src/gen \
  --go-excel_out=./server/src/gen \
  ./excelc/server/proto/*.proto
excelc code --pb_dir=./excelc/server/proto --pb_package=excel --go_out=./server/src/gen/excel

# 4. 生成客户端 GDScript protobuf 与 Excel 包装代码。
protoc -I./excelc/client/proto -I./protobuf/include \
  --include_imports \
  --retain_options \
  --gdscript_out=./client/script/gen \
  --gdscript_opt=string_as_string_name=true \
  --gdscript-excel_out=./client/script/gen \
  ./excelc/client/proto/*.proto
excelc code --pb_dir=./excelc/client/proto --pb_package=excel --gdscript_out=./client/script/gen/excel

# 5. 导出服务端表格数据。
excelc data \
  --excel_files=./config/excel/Consts.xlsx \
  --excel_dir=./config/excel \
  --pb_dir=./excelc/server/proto \
  --pb_package=excel \
  --targets=s \
  --json_out=./server/res/excel \
  --binary_out=./server/res/excel

# 6. 导出客户端分块二进制表数据。
excelc data \
  --excel_files=./config/excel/Consts.xlsx \
  --excel_dir=./config/excel \
  --pb_dir=./excelc/client/proto \
  --pb_package=excel \
  --targets=c \
  --binary_out=./client/res/excel \
  --binary_chunked=true
```

`--targets` 不只是用来区分服务端 / 客户端输出，它还会控制 Excel 导出时的列可见性。只有标记给当前 target 的字段，才会进入生成出来的 `.proto`、查询代码和最终导出的表数据。也就是说，服务端专用列可以在客户端 schema 和客户端数据中直接裁掉，客户端专用列也可以同样不进入服务端产物。

`--pb_unique_index_as` 用来控制唯一索引在导出 protobuf schema 中使用什么结构形式。不同取值会生成不同的索引消息结构，并进一步影响生成出来的查询代码以及表数据加载后的运行时内存特征。这个示例里刻意把两端分开：服务端保留 `hash_unique_index`，让生成的 Go 查询代码走哈希索引；客户端改用 `sorted_unique_index`，主要是为了降低 Godot 侧索引占用内存。生成的 GDScript Excel 包装器会对 `SortedUniqueIndex.Values` 做二分查找，并且在需要时仍可配合 `*.bin.idx` / `*.bin.chk_*` 读取分块数据。

- `hash_unique_index` 会把唯一索引生成为哈希索引结构。生成的查询代码可以基于它做按 key 的查找，对服务端这类常驻内存的数据访问场景通常更合适。
- `sorted_unique_index` 会把唯一索引生成为有序数组。查询时改为二分查找，虽然不是直接哈希访问，但通常比哈希索引占用更少内存，更适合内存受限的客户端。
- 实践上，`hash_unique_index` 更适合作为服务端查询代码的默认选择；`sorted_unique_index` 更适合 Godot 客户端这种需要控制表索引内存占用的场景。

`--binary_chunked` 用来控制 `excelc data` 导出二进制表数据时的写出方式。开启后，表行数据会拆成 `*.bin.chk_*` 分块文件，并配套生成一个 `*.bin.idx` 索引文件，适合客户端避免一次性加载整份大二进制表数据。它只影响导出的二进制布局，不会改变生成出来的 protobuf schema。

Excel 工作簿约定：

- 可选的 `@types` 页签用于声明当前工作簿里可复用的结构和枚举类型。各个表页在定义字段类型时，就可以直接引用这些自定义类型，而不只是内置标量类型。
- `@types` 页签里的 `Meta` 实际有效的主要是 `separator` 和 `scope`。`unique_index`、`hash_unique_index`、`sorted_unique_index` 这类索引参数只对表页字段列有意义，不用于 `@types` 里的结构或枚举声明。
- 表页里凡是列名首字符不是字母的列，`excelc` 都会忽略。实际使用时，通常会把 `#` 开头的列当作注释列，用来写说明、示例或仅供编辑查看的辅助信息，这些列不会进入生成的 schema，也不会进入导出的表数据。

表页数据配置：

- `ExampleCN.xlsx` 和 `ExampleEN.xlsx` 使用的是同一套表页结构：第 1 行写字段名，第 2 行写字段类型，第 3 行写字段 `Meta`，第 4 行写注释，真正的数据从第 5 行开始。
- 标量类型直接写单元格值即可，例如 `1`、`3.14`、`true`、`HelloWorld`。
- `bytes` 类型写 Base64 文本，示例里的 `SGVsbG9Xb3JsZA==` 就是这种格式。
- 枚举类型可以写枚举数值、枚举名或枚举别名。示例里三种形式都有，比如 `1`、`EnumB` / `A`，以及中文别名 `枚举值B`。
- 重复标量字段按分隔符写入，默认分隔符是 `,`，所以示例里会写成 `1,2,3,4,5`。如果在 `Meta` 里配置了 `separator=|`，同类字段也可以写成 `HelloWorld|HAHAHAHAHA`。
- 对象类型单元格使用 YAML 风格的映射写法，例如 `A: 1, B: HelloWorld, C: [1, 2, 3]`。对象字段既可以用字段名，也可以用别名，例如 `A` / `字段A`、`FieldA`。
- 重复对象字段既支持完整的 YAML 数组写法，例如 `[{A: 1, B: HelloWorld}, {A: 2, B: HAHAHAHAHA}]`，也支持在配置了自定义分隔符后写成 `A: 1, B: HelloWorld | A: 2, B: HAHAHAHAHA` 这种分隔形式。
- `map` 类型单元格同样使用 YAML 风格映射写法，例如 `0: HelloWorld, 1: HAHAHAHAHA`，或者 `0: {A: 1, B: HelloWorld}, 1: {A: 3, B: HAHAHAHAHA}`。

Excel 列参数：

- 字段级参数写在表头里的 `Meta` 设置行中，表头名也兼容 `元数据` / `特性`。每个字段列都在这一行对应的单元格里填写自己的参数，格式使用 query-string 风格，例如 `scope=c&sorted_unique_index=1` 或 `separator=|`。
- `scope`：可重复的 target 可见性标记。它和 `--targets` 一起生效；没有配置 `scope` 的列默认对所有 target 可见。需要同时命中多个 target 时，重复填写这个键即可，例如 `scope=c&scope=s` 或 `scope=client&scope=editor`。
- `separator`：重复字段或 map 类单元格值的分隔符，默认是 `,`。
- `unique_index`：可重复的整数索引 tag，用来声明唯一索引分组，最终导出成哈希索引还是有序索引取决于 `--pb_unique_index_as`。
- `hash_unique_index`：可重复的整数索引 tag，强制对应唯一索引分组使用哈希索引结构。
- `sorted_unique_index`：可重复的整数索引 tag，强制对应唯一索引分组使用有序索引结构。
- 单列唯一索引的配置方式就是把一个 tag 配到一个字段上，例如 `unique_index=1` 或 `hash_unique_index=1`。
- 复合唯一索引的配置方式是让多个字段复用同一个 tag。例如 `role_id` 写 `hash_unique_index=1`，`level` 也写 `hash_unique_index=1`，那么这两个字段会共同组成 `(role_id, level)` 这个复合唯一索引。
- 同一张表可以同时配置多个唯一索引，只要使用不同的 tag 分组即可。例如 `id -> hash_unique_index=1`，`name -> sorted_unique_index=2`，`type + sub_type -> sorted_unique_index=3`。
- 同一个字段也可以同时参与多个索引，只需要在自己的 meta 单元格里重复填写多个 tag，例如 `hash_unique_index=1&hash_unique_index=2`，这样这个字段就会同时进入两个索引分组。
- 同一个 tag 不能同时出现在 `hash_unique_index` 和 `sorted_unique_index` 中。

## 目录说明
| 路径 | 职责 |
| --- | --- |
| [`./addins/goscr`](./addins/goscr) | 服务级 Go 脚本 add-in、脚本化实体 / 组件辅助与生命周期桥接 |
| [`./addins/goscr/dynamic`](./addins/goscr/dynamic) | 动态脚本加载、工程 / 方案管理与热更新支持 |
| [`./addins/goscr/fwlib`](./addins/goscr/fwlib) | 向脚本运行时导出的 `core`、`framework`、`scaffold` 符号库 |
| [`./addins/propview`](./addins/propview) | 运行时属性视图、属性同步、序列化和复制辅助 |
| [`./tools/excelc`](./tools/excelc) | Excel 结构生成、访问代码生成和数据导出 CLI |
| [`./tools/excelc/excelutils`](./tools/excelc/excelutils) | 供生成代码使用的哈希 / 索引转换、表加载与查找辅助 |
| [`./tools/propc`](./tools/propc) | 属性同步代码生成器 |
| [`./tools/protoc-gen-go-excel`](./tools/protoc-gen-go-excel) | 面向 Excel 表访问的 Go Protobuf 插件 |
| [`./tools/protoc-gen-go-structure`](./tools/protoc-gen-go-structure) | 面向深拷贝辅助的 Go Protobuf 插件 |
| [`./tools/protoc-gen-go-variant`](./tools/protoc-gen-go-variant) | 面向 GAP variant 集成的 Go Protobuf 插件 |
| [`./tools/protoc-gen-gdscript`](./tools/protoc-gen-gdscript) | 面向消息类型与序列化逻辑的 GDScript Protobuf 插件 |
| [`./tools/protoc-gen-gdscript/libs`](./tools/protoc-gen-gdscript/libs) | 生成 `*.pb.gd` 文件所依赖的 Godot protobuf 运行时脚本目录 |
| [`./tools/protoc-gen-gdscript-excel`](./tools/protoc-gen-gdscript-excel) | 面向 Excel 表包装器的 GDScript Protobuf 插件 |
| [`./tools/protoc-gen-gdscript-excel/libs`](./tools/protoc-gen-gdscript-excel/libs) | 生成 `*.excel.gd` 包装器所依赖的 Godot Excel 运行时脚本目录 |

## 工具链说明
- `tools/excelc` 基于 Cobra / Viper，拆分为 `proto`、`code`、`data` 三个子命令。
- `tools/propc` 读取 Go 声明文件，并在旁边生成对应的 `*.sync.gen.go`。
- `tools/excelc/examples` 提供了 Excel 配表流水线的示例工作簿。

## 安装
如果需要在业务代码中直接引用 add-in 包，可安装整个模块：

```bash
go get git.golaxy.org/scaffold@latest
```

如果只需要命令行工具，可按需单独安装：

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
