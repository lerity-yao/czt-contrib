# cztctl

go-zero 官方 goctl 的增强版代码生成工具，作为独立模块发布。在保持与 goctl 兼容的基础上，扩展了 Swagger 文档生成、RabbitMQ 消费者服务生成、分布式定时任务服务生成等能力。

cztctl 语法高亮插件 [cztctl-intellij](https://github.com/lerity-yao/cztctl-intellij)

插件进入 goland 插件市场搜索 cztctl，安装即可使用

## 安装

```bash
go install github.com/lerity-yao/czt-contrib/cztctl@latest
```

验证安装：

```bash
cztctl --version
# cztctl version 1.10.2 linux/amd64 (go-zero v1.10.0)
```

## 命令一览

```
cztctl api swagger    从 .api 文件生成 Swagger 文档
cztctl api cron       从 .cron 文件生成分布式定时任务服务
cztctl api rabbitmq   从 .rabbitmq 文件生成 RabbitMQ 消费者服务
cztctl env            查看或编辑 cztctl 环境变量
```

## api swagger

从 go-zero 标准 `.api` 文件生成 Swagger 2.0 文档，增强支持 `validate` tag 自动注释和字段头部多行注释解析。

```bash
cztctl api swagger -api user.api -dir . -filename user-api.json
```

| 参数 | 说明 |
|---|---|
| `-api` | .api 文件路径 |
| `-dir` | 输出目录 |
| `-filename` | 输出文件名 |
| `--yaml` | 输出 YAML 格式 |

### info 块支持的属性

在 `.api` 文件的 `info()` 块中，以下属性会直接映射到 Swagger 文档（与 go-zero 一致）：

| 属性 | 说明 | 示例 |
|---|---|---|
| `title` | 文档标题 | `title: "用户服务"` |
| `description` | 文档描述 | `description: "用户相关接口"` |
| `version` | API 版本号 | `version: "1.0"` |
| `termsOfService` | 服务条款 URL | `termsOfService: "https://..."` |
| `contactName` | 联系人姓名 | `contactName: "张三"` |
| `contactURL` | 联系人 URL | `contactURL: "https://..."` |
| `contactEmail` | 联系人邮箱 | `contactEmail: "a@b.com"` |
| `licenseName` | 许可证名称 | `licenseName: "MIT"` |
| `licenseURL` | 许可证 URL | `licenseURL: "https://..."` |
| `host` | API 主机地址 | `host: "api.example.com"` |
| `basePath` | API 基础路径 | `basePath: "/v1"` |
| `schemes` | 协议（逗号分隔） | `schemes: "https,http"` |
| `produces` | 响应 Content-Type | `produces: "application/json"` |
| `consumes` | 请求 Content-Type | `consumes: "application/json"` |
| `useDefinitions` | 是否使用 `$ref` 引用 | `useDefinitions: "true"` |
| `wrapCodeMsg` | 是否包装 code/msg 响应 | `wrapCodeMsg: "true"` |
| `bizCodeEnumDescription` | 业务码枚举描述字段名 | `bizCodeEnumDescription: "business code"` |
| `externalDocsDescription` | 外部文档描述 | `externalDocsDescription: "详细文档"` |
| `externalDocsURL` | 外部文档 URL | `externalDocsURL: "https://..."` |

### @server 支持的注解

在 `.api` 文件的 `@server()` 块中，以下属性会影响 Swagger 生成：

| 属性 | 说明 |
|---|---|
| `tags` | 分组标签，对应 Swagger tags |
| `summary` | 接口摘要 |
| `prefix` | 路由前缀，拼接到路径前 |
| `group` | 代码分组目录 |
| `deprecated` | 标记接口已废弃 |
| `operationId` | 自定义 operationId |
| `authType` | 认证类型 |

## api cron

从 `.cron` 文件生成基于 [czt-contrib/cron](https://github.com/lerity-yao/czt-contrib/tree/main/cron) 分布式定时任务框架的定时任务服务。支持内部定时任务（`@cron`）和外部触发任务两种模式。

```bash
cztctl api cron -api task.cron -dir ./output
```

| 参数 | 说明 |
|---|---|
| `-api` | .cron 文件路径 |
| `-dir` | 输出目录 |
| `--style` | 文件命名风格，默认 `gozero` |
| `--remote` | 远程模板 Git 仓库地址 |
| `--branch` | 远程模板分支 |
| `--home` | 本地模板目录 |

### 生成的目录结构

```
output/
├── etc/
│   └── usercron.yaml         配置文件
├── internal/
│   ├── config/
│   │   └── config.go         配置结构体
│   ├── handler/              handler 层（任务注册）
│   │   └── user/             按 group 分目录
│   ├── logic/                logic 层（业务代码写这里）
│   │   └── user/
│   ├── svc/
│   │   └── servicecontext.go 服务上下文
│   ├── types/
│   │   └── types.go          请求/响应类型
│   └── worker/
│       └── worker.go         任务注册（cron 调度 + asynq handler 绑定）
└── usercron.go               主入口
```

rabbitmq 生成结构与 cron 基本一致，区别在于 worker 层处理的是 MQ 消费而非 cron 调度。

### .cron 文件语法

`.cron` 文件采用类 go-zero api 的 DSL 语法，由以下顶层声明组成：

#### syntax（必填）

声明语法版本，当前固定为 `"v1"`。

```
syntax = "v1"
```

#### info（可选）

描述服务的元信息，支持任意 key-value 对：

```
info (
    title: "用户定时任务服务"
    desc: "处理用户相关的定时与异步任务"
    version: "1.0"
    author: "张三"
    email: "zhangsan@example.com"
)
```

#### import（可选）

导入其他 `.cron` 文件中的类型定义，支持单行和分组两种写法：

```
import "common.cron"

import (
    "user.cron"
    "order.cron"
)
```

#### type（可选）

定义请求参数结构体，语法与 go-zero api 的 type 完全一致，支持基本类型、切片、map、指针、嵌套结构体、struct tag：

```
type CleanReq {
    UserId int64  `json:"userId"`
    Action string `json:"action"`
}

type (
    ArchiveReq {
        StartDate string `json:"startDate"`
        EndDate   string `json:"endDate"`
    }

    NotifyReq {
        Email   string   `json:"email"`
        Content string   `json:"content"`
        Tags    []string `json:"tags"`
    }
)
```

#### @server（可选）

为 service 块设置元数据，支持以下字段：

| 字段 | 说明 | 示例 |
|---|---|---|
| `tags` | Swagger 分组标签 | `tags: "用户管理"` |
| `summary` | 服务摘要 | `summary: "用户定时任务"` |
| `description` | 服务描述 | `description: "详细说明"` |
| `group` | 代码分组目录 | `group: user` |
| `middleware` | 中间件（逗号分隔多个） | `middleware: Auth,Log` |

```
@server (
    tags: "用户管理"
    summary: "用户相关定时任务"
    group: user
    middleware: LogMiddleware
)
```

#### service（必填）

定义服务和任务列表。每个任务项由以下注解组成：

| 注解 | 必填 | 说明 |
|---|---|---|
| `@doc` | 否 | 任务文档，支持字符串和 KV 两种写法 |
| `@cron` | 否 | cron 表达式，有则为内部定时任务，无则为外部触发任务 |
| `@cronRetry` | 否 | 失败重试次数，整数 |
| `@handler` | **是** | 任务名称，即 asynq task type |
| 路由行 | **是** | 任务标识，可选带请求参数 `TaskName(ReqType)` |

**路由名规则：**

路由名由标识符（字母、数字、下划线、`$`）组成，支持 `-`（横杠）和 `:`（冒号）作为分隔符：

```
CleanUserTask           // 纯标识符
sync-order              // 横杠分隔
email:send              // 冒号分隔
data-sync:daily         // 混合分隔
```

**@doc 两种写法：**

```
// 字符串写法
@doc "简单的文档注释"

// KV 写法
@doc(
    summary: "数据归档任务"
    description: "每天凌晨归档历史数据"
)
```

**完整示例：**

```
syntax = "v1"

info (
    title: "用户定时任务服务"
    version: "1.0"
)

type (
    CleanReq {
        UserId int64 `json:"userId"`
    }
)

@server (
    tags: "用户管理"
    group: user
)
service userCron {
    // 内部定时任务：每分钟执行，失败重试 3 次
    @doc "清理过期用户数据"
    @cron "*/1 * * * *"
    @cronRetry 3
    @handler CleanUserJob
    CleanUserTask(CleanReq)

    // 内部定时任务：无参数
    @doc(
        summary: "数据归档"
        description: "每天凌晨 2 点归档"
    )
    @cron "0 2 * * *"
    @handler ArchiveDataJob
    ArchiveDataTask

    // 外部触发任务：无 @cron，由业务代码调用 Add 触发
    @doc "发送邮件通知"
    @handler SendEmailJob
    SendEmailTask(CleanReq)
}
```

**任务类型区分：**

- 有 `@cron` → 内部定时任务，框架按 cron 表达式自动调度
- 无 `@cron` → 外部触发任务，通过 `asynq.Client.Enqueue` 手动触发（支持延时执行、立即执行）

## api rabbitmq

从 `.rabbitmq` 文件生成基于 [czt-contrib/mq/rabbitmq](https://github.com/lerity-yao/czt-contrib/tree/main/mq/rabbitmq) 分布式 RabbitMQ 消费者服务框架的 RabbitMQ 消费者服务。

```bash
cztctl api rabbitmq -api order.rabbitmq -dir ./output
```

| 参数 | 说明 |
|---|---|
| `-api` | .rabbitmq 文件路径 |
| `-dir` | 输出目录 |
| `--style` | 文件命名风格，默认 `gozero` |
| `--remote` | 远程模板 Git 仓库地址 |
| `--branch` | 远程模板分支 |
| `--home` | 本地模板目录 |

### .rabbitmq 文件语法

`.rabbitmq` 文件与 `.cron` 文件共享相同的基础语法（syntax / info / import / type / @server），区别在于 service 块内的任务定义。

#### service（必填）

每个任务项由以下注解组成：

| 注解 | 必填 | 说明 |
|---|---|---|
| `@doc` | 否 | 消费者文档，支持字符串和 KV 两种写法 |
| `@handler` | **是** | 消费者处理器名称 |
| 路由行 | **是** | 队列名称（点分隔标识符），可选带消息参数 |

**队列名称规则：**

路由名由标识符（字母、数字、下划线、`$`）组成，支持 `.`（点号）和 `-`（横杠）作为分隔符，对应 RabbitMQ 的队列名称：

```
order.created           // 点号分隔
payment.refund.success  // 多段点号
payment-refund          // 横杠分隔
order.pay-callback      // 混合分隔
```

**完整示例：**

```
syntax = "v1"

info (
    title: "订单消息队列消费者"
    desc: "处理订单相关的 MQ 事件"
    version: "1.0"
)

type (
    OrderCreatedEvent {
        OrderId int64   `json:"orderId"`
        Amount  float64 `json:"amount"`
    }
)

@server (
    tags: "订单管理"
    summary: "订单事件消费者"
    group: order
    middleware: LogMiddleware
)
service order-mq {
    // 无消息参数的消费者
    @doc "订单创建事件"
    @handler OrderCreatedConsumer
    order.created

    // 带消息参数的消费者
    @doc(
        summary: "支付成功事件"
        description: "处理支付成功后的后续逻辑"
    )
    @handler PaymentSuccessConsumer
    payment.success(OrderCreatedEvent)

    // 简单字符串文档
    @doc "用户注册事件"
    @handler UserRegisteredConsumer
    user.registered
}
```

### .cron 与 .rabbitmq 语法对比

| 特性 | .cron | .rabbitmq |
|---|---|---|
| syntax / info / import / type | 完全相同 | 完全相同 |
| @server | 完全相同 | 完全相同 |
| @doc | 完全相同 | 完全相同 |
| @cron | 支持 | 不支持 |
| @cronRetry | 支持 | 不支持 |
| @handler | task type 名称 | 消费者名称 |
| 路由行格式 | `TaskName[(ReqType)]` | `queue.name[(MsgType)]` |
| 路由名分隔符 | `-`（横杠）、`:`（冒号） | `.`（点号）、`-`（横杠） |
| 路由行命名 | 标识符，支持横杠/冒号分隔 | 标识符，支持点号/横杠分隔 |

## 远程模板

cron 和 rabbitmq 命令支持通过 `--remote` 和 `--branch` 拉取远程 Git 仓库中的自定义模板，覆盖内置模板：

```bash
cztctl api cron \
  -api task.cron \
  -dir ./output \
  --remote https://your-git-repo/goctl-template.git \
  --branch dev
```

模板文件会被 clone 到 `~/.cztctl/.git/` 目录下，后续模板加载优先使用远程模板。

## 文件命名风格

通过 `--style` 控制生成文件的命名风格：

| 风格 | 示例 |
|---|---|
| `gozero`（默认） | `servicecontext.go` |
| `go_zero` | `service_context.go` |
| `goZero` | `serviceContext.go` |

## env

查看或编辑 cztctl 环境变量。环境配置持久化在 `~/.cztctl/env` 文件中。

### 查看环境变量

```bash
cztctl env
```

输出示例：

```
CZTCTL_OS=linux
CZTCTL_ARCH=amd64
CZTCTL_HOME=/home/user/.cztctl
CZTCTL_CACHE=/home/user/.cztctl/cache
CZTCTL_EXPERIMENTAL=off
CZTCTL_VERSION=1.10.2
```

### 编辑环境变量

```bash
cztctl env -w KEY=VALUE
```

支持同时设置多个值：

```bash
cztctl env -w CZTCTL_EXPERIMENTAL=on -w CZTCTL_HOME=/custom/path
```

### 支持的变量

| 变量 | 说明 | 默认值 |
|---|---|---|
| `CZTCTL_OS` | 操作系统（只读） | `runtime.GOOS` |
| `CZTCTL_ARCH` | 系统架构（只读） | `runtime.GOARCH` |
| `CZTCTL_HOME` | cztctl 主目录 | `~/.cztctl` |
| `CZTCTL_CACHE` | 缓存目录 | `~/.cztctl/cache` |
| `CZTCTL_VERSION` | 当前版本（只读） | 构建版本号 |
| `CZTCTL_EXPERIMENTAL` | 实验性功能开关 | `off` |

### CZTCTL_EXPERIMENTAL

控制 DSL 解析器选择：

- `off`（默认）：使用 ANTLR4 解析器
- `on`：使用手写递归下降解析器

```bash
# 切换到手写解析器
cztctl env -w CZTCTL_EXPERIMENTAL=on

# 切换回 ANTLR4 解析器
cztctl env -w CZTCTL_EXPERIMENTAL=off
```

## 版本规则

版本号格式：`v<go-zero主版本>.<微版本>`

- 前两段（如 `1.10`）对应所依赖的 go-zero 版本
- 末段为 cztctl 自身递增版本号
- `--version` 输出自动关联 go.mod 中的 go-zero 版本

## License

同 go-zero，MIT License。
