# cztctl 维护更新说明

## 项目结构

```
cztctl/
├── cztctl.go                     主入口
├── cmd/                          CLI 命令注册
├── api/
│   ├── cmd.go                    api 子命令注册（swagger/cron/rabbitmq）
│   ├── swagger/                  Swagger 文档生成（依赖 goctl api/spec + api/parser）
│   ├── cron/                     Cron 定时任务服务生成
│   ├── rabbitmq/                 RabbitMQ 消费者服务生成
│   ├── spec/                     .cron/.rabbitmq 的 AST 规格定义（ApiSpec/Route/Service 等）
│   ├── parser/                   .cron/.rabbitmq 文件解析器入口
│   │   ├── parser.go             双 parser 路由（默认 ANTLR4，实验版手写递归下降）
│   │   └── g4/                   ANTLR4 parser 实现
│   │       ├── CztctlLexer.g4    Lexer 语法文件
│   │       ├── CztctlParser.g4   Parser 语法文件
│   │       ├── ast/              AST 构建（Visitor 遍历 parse tree → spec 结构）
│   │       ├── gen/cztctl/       ANTLR4 生成的 Go 代码（勿手动修改）
│   │       └── test/             g4 parser 测试用例
│   ├── gogen/                    代码生成工具函数
│   └── apiutil/                  API 工具函数
├── pkg/
│   ├── parser/extension/         实验版手写递归下降 parser（备用）
│   │   ├── parser/               解析器入口
│   │   ├── ast/                  AST 节点定义
│   │   ├── g4/                   从 goctl 复制的 g4 相关代码
│   │   ├── scanner/              词法扫描器
│   │   └── token/                Token 定义
│   └── golang/                   Go 代码格式化工具
├── config/                       文件命名风格配置
├── internal/
│   ├── version/                  版本管理
│   ├── cobrax/                   Cobra 命令封装
│   └── flags/                    Flag 描述配置
├── env/                          env 子命令（查看/编辑环境变量）
├── util/                         工具函数（git/模板/路径/格式化/环境配置等）
│   ├── env/                      环境配置（init/Print/Get/WriteEnv/UseExperimental）
├── test/                         测试文件（test.cron / test.rabbitmq）
└── vars/                         全局常量
```

## 架构要点

### 双 parser 架构

`api/parser/parser.go` 中的 `Parse()` 函数通过 `env.UseExperimental()` 路由到两套 parser：

| parser | 路径 | 使用条件 |
|---|---|---|
| ANTLR4 parser（默认） | `api/parser/g4/` | `CZTCTL_EXPERIMENTAL=off`（默认值） |
| 手写递归下降 parser（实验版） | `pkg/parser/extension/` | `CZTCTL_EXPERIMENTAL=on` |

通过 `cztctl env -w CZTCTL_EXPERIMENTAL=on` 切换到手写 parser，通过 `cztctl env -w CZTCTL_EXPERIMENTAL=off` 切回 ANTLR4。

`util/env/env.go` 中的 `init()` 在启动时读取 `~/.cztctl/env` 文件加载配置，`WriteEnv()` 负责持久化。env 管理使用轻量级的 `orderedEnv`（有序 map），不依赖 `tools/cztctl/pkg/collection/sortedmap`。

两套 parser 必须共存，不能删除任何一套。

### 两个不同的 spec 包

项目中存在两个 `spec` 包，用途完全不同：

| 包 | 路径 | 用途 |
|---|---|---|
| cztctl spec | `api/spec/spec.go` | .cron/.rabbitmq 的 AST 定义，cron/rabbitmq 生成器使用 |
| goctl spec | `goctl/api/spec` | .api 的 AST 定义，swagger 生成器使用 |

swagger 模块直接依赖 goctl 的 spec（`github.com/zeromicro/go-zero/tools/goctl/api/spec`），因为它处理标准 `.api` 文件。cron/rabbitmq 模块使用 cztctl 自己的 spec（`github.com/lerity-yao/czt-contrib/cztctl/api/spec`），因为它处理扩展的 `.cron`/`.rabbitmq` 文件。

### .g4 语法共享与语义隔离

`.cron` 和 `.rabbitmq` 共用同一套 .g4 语法文件（`CztctlLexer.g4` / `CztctlParser.g4`）。语法层面两者完全一致，区分由 `parser.go` 的 `fillService()` 方法在语义层完成：

```go
isCron := strings.HasSuffix(filename, ".cron")
if isCron {
    // 填充 route.Cron、route.CronRetry
} else {
    // 填充 route.Queue
}
```

即：按文件后缀决定填充哪些字段。

### baseparser.go 是手写文件

`api/parser/g4/gen/cztctl/baseparser.go` 是**手写**的辅助文件（包含 `IsBasicType()`、`IsGolangKeyWord()`、`checkVersion()`、import 路径校验等），**不是** ANTLR 生成的。它与 ANTLR 生成的 6 个 `cztctlparser_*.go` 文件放在同一目录下。

**重新生成 ANTLR 代码时，不能删除 `baseparser.go`。** 只删除 `cztctlparser_*.go` 文件即可。

## ANTLR4 Parser 维护指南

### 环境要求

| 依赖 | 版本 | 说明 |
|---|---|---|
| ANTLR jar | **4.7.2** | 必须使用此版本，不可用 4.10+ |
| Java 运行时 | JDK 8+ | 若系统无 JDK，可用 GoLand 自带 JBR：`/path/to/goland/jbr/bin/java` |
| Go antlr runtime | `github.com/zeromicro/antlr v1.0.0` | 非官方 `github.com/antlr/antlr4/runtime/Go/antlr` |

### 踩坑记录

#### 1. ANTLR 版本不兼容（最关键）

`zeromicro/antlr v1.0.0` 是基于 ANTLR 4.7.x 风格的 Go runtime，只提供 `DeserializeFromUInt16([]uint16)` 方法。

而 ANTLR 4.10+ 生成的代码使用 `Deserialize([]int32)` + `staticData` / `sync.Once` 初始化模式，与此 runtime **完全不兼容**，编译会报 `deserializer.Deserialize undefined`。

**结论：必须用 ANTLR 4.7.2 的 jar 来生成代码。** 下载地址：
```
https://www.antlr.org/download/antlr-4.7.2-complete.jar
```

#### 2. 双 .g4 文件生成冲突

同时传入 `CztctlLexer.g4` 和 `CztctlParser.g4` 会生成两个 lexer 文件：
- `cztctl_lexer.go`（独立 lexer，来自 CztctlLexer.g4）
- `cztctlparser_lexer.go`（parser 附带 lexer，来自 CztctlParser.g4）

两者定义了相同的包级变量（`serializedLexerAtn`、`lexerAtn`、`lexerDecisionToDFA`、`init()`），导致编译冲突。

**解决方案：** 生成后删除 `cztctl_lexer.go`，只保留 `cztctlparser_lexer.go`。

#### 3. kvValue 是 parser rule，不是 token

在 `CztctlParser.g4` 中：
```
kvLit:  key=ID ':' value=kvValue;
kvValue: STRING | RAW_STRING | INT | ID ((',' | '-') ID)*;
```

`kvValue` 是一条 parser rule（非 lexer token），因此 `ctx.GetValue()` 返回的是 `IKvValueContext`（rule context），**不是** `antlr.Token`。不能传给 `newExprWithToken()`，必须用 `newExprWithText()` + `GetStart()/GetStop()` 来获取位置信息。

#### 4. ANTLR 元数据文件无需提交

生成时会产生 `.interp` 和 `.tokens` 文件（共 6 个），这些是 ANTLR 调试用的元数据，**应删除，不提交到仓库**。

### 重新生成 ANTLR 代码的命令

```bash
cd api/parser/g4

# 先删除旧的生成文件（保留手写的 baseparser.go）
rm -f gen/cztctl/cztctlparser_*.go

# 生成（-o 指定输出目录，-package 指定包名）
java -jar /path/to/antlr-4.7.2-complete.jar \
  -Dlanguage=Go -visitor \
  -o gen/cztctl -package cztctl \
  CztctlLexer.g4 CztctlParser.g4

# 修复 import：zeromicro/antlr 不是官方路径
find gen/cztctl -name '*.go' -exec sed -i \
  's|github.com/antlr/antlr4/runtime/Go/antlr|github.com/zeromicro/antlr|g' {} +

# 删除冲突的独立 lexer 文件
rm -f gen/cztctl/cztctl_lexer.go

# 删除 ANTLR 元数据文件
rm -f gen/cztctl/*.interp gen/cztctl/*.tokens
```

生成后 `gen/cztctl/` 目录结构：
```
gen/cztctl/
├── baseparser.go                   手写辅助文件（勿删）
├── cztctlparser_base_listener.go   ANTLR 生成
├── cztctlparser_base_visitor.go    ANTLR 生成
├── cztctlparser_lexer.go           ANTLR 生成
├── cztctlparser_listener.go        ANTLR 生成
├── cztctlparser_parser.go          ANTLR 生成
└── cztctlparser_visitor.go         ANTLR 生成
```

## 更新 cron/rabbitmq 语法扩展流程

当需要为 `.cron` 或 `.rabbitmq` 文件新增语法（如新增 `@cronTimeout` 注解），按以下步骤操作：

### 第一步：修改 .g4 语法文件

编辑 `api/parser/g4/CztctlParser.g4`（如需新增 token 则同时编辑 `CztctlLexer.g4`）。

例如新增 `@cronTimeout`：
```
// CztctlLexer.g4 中新增 token（如果需要新关键字）
ATCRONTIMEOUT: '@cronTimeout';

// CztctlParser.g4 中新增 rule
atCronTimeout: ATCRONTIMEOUT INT;

// 在 serviceRoute 中引用
serviceRoute: ... atCronTimeout? ... ;
```

### 第二步：重新生成 Go 代码

执行上述「重新生成 ANTLR 代码的命令」中的完整流程（生成 → 替换 import → 删除冲突文件 → 删除元数据）。

### 第三步：修改 AST 层

在 `api/parser/g4/ast/service.go` 中的 `VisitServiceRoute` 方法中，处理新增的 parse tree 节点：

```go
if ctx.AtCronTimeout() != nil {
    // 提取 timeout 值并填充到 route 结构
}
```

### 第四步：修改 spec 定义

在 `api/spec/spec.go` 的 `Route` 结构体中新增字段：

```go
type Route struct {
    // ...
    CronTimeout int  // 新增
}
```

### 第五步：修改 parser.go 语义填充

在 `api/parser/parser.go` 的 `fillService()` 方法中，将新字段从中间 spec 传递到最终 spec。

### 第六步：修改代码生成模板

在 `api/cron/` 下修改生成模板，使用新字段生成对应代码。

### 第七步：编译验证

```bash
cd /path/to/czt-contrib/cztctl
go build ./...
```

确保零错误后再提交。

> **注意：** 如果实验版 parser（`pkg/parser/extension/`）仍在维护，也需要同步更新其 scanner/ast 层以支持新语法。

## Swagger 同步 go-zero 最新版本

Swagger 模块直接依赖 goctl 的 `api/spec` 和 `pkg/parser/api/parser` 包：

```go
import (
    "github.com/zeromicro/go-zero/tools/goctl/api/spec"       // 类型定义
    "github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/parser" // .api 文件解析
    "github.com/zeromicro/go-zero/tools/goctl/util/pathx"     // 路径工具
    "github.com/zeromicro/go-zero/tools/goctl/util/stringx"   // 字符串工具
)
```

当 go-zero 发布新版本后，同步步骤如下：

### 第一步：更新 go.mod 依赖

```bash
cd /path/to/czt-contrib/cztctl

# 更新 go-zero 和 goctl 到目标版本
go get github.com/zeromicro/go-zero@v1.11.0
go get github.com/zeromicro/go-zero/tools/goctl@v1.11.0
go mod tidy
```

### 第二步：检查 spec.Type 接口变更

重点关注 `goctl/api/spec` 包中的类型变化：
- `spec.Type` 接口是否新增方法
- `spec.DefineStruct`、`spec.Tags` 等结构体字段是否变化
- `spec.Route`、`spec.Group` 是否新增字段

如果有 breaking change，需要同步修改 `api/swagger/` 下的类型断言和字段引用。

### 第三步：检查 parser API 变更

确认 `goctl/pkg/parser/api/parser.Parse()` 的签名是否变化。如有变化，同步修改 `api/swagger/command.go` 中的调用。

### 第四步：编译验证

```bash
go build ./...
```

### 第五步：更新版本号

修改 `internal/version/` 中的版本号，前两段对应 go-zero 版本。
