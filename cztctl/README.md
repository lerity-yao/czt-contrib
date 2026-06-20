# cztctl

[中文](./readme-cn.md)

An enhanced code generation tool based on the official go-zero `goctl`, published as a standalone module. It maintains full compatibility with goctl while extending capabilities including Swagger documentation generation, RabbitMQ consumer service generation, and distributed cron job service generation.

Syntax highlighting plugin: [cztctl-intellij](https://github.com/lerity-yao/cztctl-intellij)

Search for `cztctl` in the GoLand plugin marketplace to install.

## Installation

```bash
go install github.com/lerity-yao/czt-contrib/cztctl@latest
```

Verify installation:

```bash
cztctl --version
# cztctl version 1.10.2 linux/amd64 (go-zero v1.10.0)
```

## Command Overview

```
cztctl api swagger    Generate Swagger documentation from a .api file
cztctl api cron       Generate a distributed cron job service from a .cron file
cztctl api rabbitmq   Generate a RabbitMQ consumer service from a .rabbitmq file
cztctl env            View or edit cztctl environment variables
cztctl rpc sdk        Generate an RPC client SDK and publish it to a standalone Git repository
```

## api swagger

Generate Swagger 2.0 documentation from a standard go-zero `.api` file, with enhanced support for automatic `validate` tag annotations and multi-line field header comment parsing.

```bash
cztctl api swagger -api user.api -dir . -filename user-api.json
```

| Flag | Description |
|---|---|
| `-api` | Path to the .api file |
| `-dir` | Output directory |
| `-filename` | Output filename |
| `--yaml` | Output in YAML format |

### Supported `info` Block Properties

The following properties in the `info()` block of a `.api` file are mapped directly to the Swagger document (consistent with go-zero):

| Property | Description | Example |
|---|---|---|
| `title` | Document title | `title: "User Service"` |
| `description` | Document description | `description: "User-related APIs"` |
| `version` | API version | `version: "1.0"` |
| `termsOfService` | Terms of service URL | `termsOfService: "https://..."` |
| `contactName` | Contact name | `contactName: "John Doe"` |
| `contactURL` | Contact URL | `contactURL: "https://..."` |
| `contactEmail` | Contact email | `contactEmail: "a@b.com"` |
| `licenseName` | License name | `licenseName: "MIT"` |
| `licenseURL` | License URL | `licenseURL: "https://..."` |
| `host` | API host address | `host: "api.example.com"` |
| `basePath` | API base path | `basePath: "/v1"` |
| `schemes` | Protocols (comma-separated) | `schemes: "https,http"` |
| `produces` | Response Content-Type | `produces: "application/json"` |
| `consumes` | Request Content-Type | `consumes: "application/json"` |
| `useDefinitions` | Whether to use `$ref` references | `useDefinitions: "true"` |
| `wrapCodeMsg` | Whether to wrap responses with code/msg | `wrapCodeMsg: "true"` |
| `bizCodeEnumDescription` | Business code enum description field name | `bizCodeEnumDescription: "business code"` |
| `externalDocsDescription` | External docs description | `externalDocsDescription: "Full docs"` |
| `externalDocsURL` | External docs URL | `externalDocsURL: "https://..."` |

### Supported `@server` Annotations

The following properties in the `@server()` block of a `.api` file affect Swagger generation:

| Property | Description |
|---|---|
| `tags` | Grouping tags, mapped to Swagger tags |
| `summary` | API summary |
| `prefix` | Route prefix, prepended to the path |
| `group` | Code grouping directory |
| `deprecated` | Marks the API as deprecated |
| `operationId` | Custom operationId |
| `authType` | Authentication type |

## api cron

Generate a cron job service based on the [czt-contrib/cron](https://github.com/lerity-yao/czt-contrib/tree/main/cron) distributed cron framework from a `.cron` file. Supports both internal scheduled tasks (`@cron`) and externally triggered tasks.

```bash
cztctl api cron -api task.cron -dir ./output
```

| Flag | Description |
|---|---|
| `-api` | Path to the .cron file |
| `-dir` | Output directory |
| `--style` | File naming style, default `gozero` |
| `--remote` | Remote template Git repository URL |
| `--branch` | Remote template branch |
| `--home` | Local template directory |

### Generated Directory Structure

```
output/
├── etc/
│   └── usercron.yaml         Configuration file
├── internal/
│   ├── config/
│   │   └── config.go         Configuration struct
│   ├── handler/              Handler layer (task registration)
│   │   └── user/             Subdirectory by group
│   ├── logic/                Logic layer (write business code here)
│   │   └── user/
│   ├── svc/
│   │   └── servicecontext.go Service context
│   ├── types/
│   │   └── types.go          Request/response types
│   └── worker/
│       └── worker.go         Task registration (cron scheduling + asynq handler binding)
└── usercron.go               Main entry point
```

The rabbitmq generated structure is essentially the same as cron, except the worker layer handles MQ consumption instead of cron scheduling.

### .cron File Syntax

The `.cron` file uses a DSL syntax similar to the go-zero api format, composed of the following top-level declarations:

#### syntax (required)

Declares the syntax version, currently fixed as `"v1"`.

```
syntax = "v1"
```

#### info (optional)

Describes service metadata, supports arbitrary key-value pairs:

```
info (
    title: "User Cron Service"
    desc: "Handles user-related scheduled and async tasks"
    version: "1.0"
    author: "John Doe"
    email: "johndoe@example.com"
)
```

#### import (optional)

Imports type definitions from other `.cron` files. Supports both single-line and grouped forms:

```
import "common.cron"

import (
    "user.cron"
    "order.cron"
)
```

#### type (optional)

Defines request parameter structs. The syntax is identical to go-zero api's type declarations and supports primitive types, slices, maps, pointers, nested structs, and struct tags:

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

#### @server (optional)

Sets metadata for the service block. Supported fields:

| Field | Description | Example |
|---|---|---|
| `tags` | Swagger grouping tags | `tags: "User Management"` |
| `summary` | Service summary | `summary: "User cron tasks"` |
| `description` | Service description | `description: "Detailed description"` |
| `group` | Code grouping directory | `group: user` |
| `middleware` | Middleware (comma-separated) | `middleware: Auth,Log` |

```
@server (
    tags: "User Management"
    summary: "User-related cron tasks"
    group: user
    middleware: LogMiddleware
)
```

#### service (required)

Defines the service and task list. Each task entry is composed of the following annotations:

| Annotation | Required | Description |
|---|---|---|
| `@doc` | No | Task documentation, supports both string and KV forms |
| `@cron` | No | Cron expression; if present, the task is an internal scheduled task; otherwise it is externally triggered |
| `@cronRetry` | No | Number of retry attempts on failure, integer |
| `@handler` | **Yes** | Task name, i.e., the asynq task type |
| Route line | **Yes** | Task identifier, optionally with a request parameter `TaskName(ReqType)` |

**Route name rules:**

Route names are composed of identifiers (letters, digits, underscores, `$`) and support `-` (hyphen) and `:` (colon) as separators:

```
CleanUserTask           // plain identifier
sync-order              // hyphen-separated
email:send              // colon-separated
data-sync:daily         // mixed separators
```

**Two forms of `@doc`:**

```
// String form
@doc "Simple documentation comment"

// KV form
@doc(
    summary: "Data archival task"
    description: "Archive historical data at midnight every day"
)
```

**Full example:**

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

**Task type distinction:**

- With `@cron` → Internal scheduled task, automatically dispatched by the framework according to the cron expression
- Without `@cron` → Externally triggered task, triggered manually via `asynq.Client.Enqueue` (supports delayed or immediate execution)

## api rabbitmq

Generate a RabbitMQ consumer service based on the [czt-contrib/mq/rabbitmq](https://github.com/lerity-yao/czt-contrib/tree/main/mq/rabbitmq) distributed RabbitMQ consumer framework from a `.rabbitmq` file.

```bash
cztctl api rabbitmq -api order.rabbitmq -dir ./output
```

| Flag | Description |
|---|---|
| `-api` | Path to the .rabbitmq file |
| `-dir` | Output directory |
| `--style` | File naming style, default `gozero` |
| `--remote` | Remote template Git repository URL |
| `--branch` | Remote template branch |
| `--home` | Local template directory |

### .rabbitmq File Syntax

The `.rabbitmq` file shares the same base syntax as `.cron` (syntax / info / import / type / @server). The difference lies in the task definitions within the service block.

#### service (required)

Each task entry is composed of the following annotations:

| Annotation | Required | Description |
|---|---|---|
| `@doc` | No | Consumer documentation, supports both string and KV forms |
| `@handler` | **Yes** | Consumer handler name |
| Route line | **Yes** | Queue name (dot-separated identifier), optionally with a message parameter |

**Queue name rules:**

Route names are composed of identifiers (letters, digits, underscores, `$`) and support `.` (dot) and `-` (hyphen) as separators, corresponding to RabbitMQ queue names:

```
order.created           // dot-separated
payment.refund.success  // multi-segment dot
payment-refund          // hyphen-separated
order.pay-callback      // mixed separators
```

**Full example:**

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

### .cron vs .rabbitmq Syntax Comparison

| Feature | .cron | .rabbitmq |
|---|---|---|
| syntax / info / import / type | Identical | Identical |
| @server | Identical | Identical |
| @doc | Identical | Identical |
| @cron | Supported | Not supported |
| @cronRetry | Supported | Not supported |
| @handler | task type name | consumer name |
| Route line format | `TaskName[(ReqType)]` | `queue.name[(MsgType)]` |
| Route name separators | `-` (hyphen), `:` (colon) | `.` (dot), `-` (hyphen) |
| Route naming | Identifier with hyphen/colon | Identifier with dot/hyphen |

## Remote Templates

The `cron` and `rabbitmq` commands support pulling custom templates from a remote Git repository via `--remote` and `--branch`, overriding the built-in templates:

```bash
cztctl api cron \
  -api task.cron \
  -dir ./output \
  --remote https://your-git-repo/goctl-template.git \
  --branch dev
```

Template files are cloned into `~/.cztctl/.git/`. Subsequent template loading will prefer remote templates.

## File Naming Styles

Control the naming style of generated files via `--style`:

| Style | Example |
|---|---|
| `gozero` (default) | `servicecontext.go` |
| `go_zero` | `service_context.go` |
| `goZero` | `serviceContext.go` |

## rpc sdk

Automatically generate RPC service client code as a standalone Go module and push it to a specified Git repository. Consumers can import it via `go get` without worrying about proto definitions or code generation.

```bash
cztctl rpc sdk --proto proto/order.proto --repo https://gitlab.ddtz.com/rpc-sdk/order-sdk.git
```

| Flag | Short | Required | Default | Description |
|---|---|---|---|---|
| `--proto` | - | Yes | - | Path to the proto file (relative or absolute) |
| `--repo` | - | Yes | - | Full git URL of the SDK repository (http:// or https://) |
| `--repo-user` | - | No | `cztctl-bot` | Repository authentication username |
| `--repo-token` | - | No | Built-in default | Repository authentication credential (GitLab Access Token) |
| `--remote` | - | No | empty | goctl remote template git URL |
| `--style` | - | No | `gozero` | Naming style |
| `--tag` | - | No | Auto-incremented | Version number (SemVer format, e.g. v1.0.0) |
| `--branch` | - | No | empty | goctl remote template branch |
| `-m` | `--multiple` | No | `false` | Multi-service mode; required if the rpc service uses multiple mode |
| `--repo-branch` | - | No | `main` | SDK repository Git branch name |
| `--goproxy` | - | No | System default | Go module proxy address; the tool runs `go mod tidy`, pass a proxy if there are network issues |

### Workflow

1. Clone the existing SDK repository (or initialize a new one)
2. Clean up old client code
3. Initialize go.mod (if it does not exist)
4. Recursively copy proto files (including dependency protos)
5. Auto-generate `.kong.proto` (Kong gRPC-gateway HTTP annotation variant)
6. Call goctl to generate client code
7. Clean up server-side code (keep only the `client/` directory)
8. Run `go mod tidy` to tidy dependencies
9. Git commit, tag, and push to the remote repository

### Automatic Kong gRPC-gateway Proto Generation

When `cztctl rpc sdk` is executed, a `.kong.proto` file with the same name as the proto is automatically generated in the `_sdk/` directory, used for Kong gateway gRPC-gateway routing configuration.

**Generation rules:**

- Automatically adds `import "google/api/annotations.proto"`
- Generates `option (google.api.http)` annotations for each rpc method
- HTTP path: `/{ServiceName}/{RpcMethodName}`
- HTTP method is uniformly `POST` with `body: "*"`

**Example:**

Original `vehicle.proto`:

```proto
service SfVehicle {
  rpc SfTaskCreate(SfTaskCreateReq) returns (SfTaskCreateRes);
}
```

Generated `vehicle.kong.proto`:

```proto
import "google/api/annotations.proto";

service SfVehicle {
  rpc SfTaskCreate(SfTaskCreateReq) returns (SfTaskCreateRes) {
    option (google.api.http) = { post: "/SfVehicle/SfTaskCreate" body: "*" };
  }
}
```

### Version Number Rules

- First release default: `v1.0.0`
- Subsequent patch auto-increment: `v1.0.0` → `v1.0.1` → `v1.0.2` → ...
- Auto-carry minor when patch reaches 99: `v1.0.99` → `v1.1.0`
- Can be manually specified via `--tag` (must be greater than the current latest version)

### Usage Examples

**Minimal usage:**

```bash
cztctl rpc sdk \
  --proto proto/tax_invoice.proto \
  --repo https://gitlab.ddtz.com/rpc-sdk/tax-invoice-sdk.git
```

**Full example:**

```bash
cztctl rpc sdk \
  --proto proto/order.proto \
  --repo https://gitlab.ddtz.com/rpc-sdk/order-sdk.git \
  --style gozero \
  --branch develop \
  -m \
  --repo-branch main \
  --goproxy https://goproxy.cn,direct
```

### SDK Consumer Usage

```bash
# Pull the SDK
go get gitlab.ddtz.com/rpc-sdk/order-sdk@latest

# Or specify a version
go get gitlab.ddtz.com/rpc-sdk/order-sdk@v1.0.2
```

```go
import "gitlab.ddtz.com/rpc-sdk/order-sdk/client/order"

client := order.NewOrder(zrpc.MustNewClient(conf.OrderRpc))
resp, err := client.GetOrder(ctx, &order.GetOrderRequest{Id: 123})
```

## env

View or edit cztctl environment variables. Environment configuration is persisted in `~/.cztctl/env`.

### View Environment Variables

```bash
cztctl env
```

Sample output:

```
CZTCTL_OS=linux
CZTCTL_ARCH=amd64
CZTCTL_HOME=/home/user/.cztctl
CZTCTL_CACHE=/home/user/.cztctl/cache
CZTCTL_EXPERIMENTAL=off
CZTCTL_VERSION=1.10.2
```

### Edit Environment Variables

```bash
cztctl env -w KEY=VALUE
```

Multiple values can be set at the same time:

```bash
cztctl env -w CZTCTL_EXPERIMENTAL=on -w CZTCTL_HOME=/custom/path
```

### Supported Variables

| Variable | Description | Default |
|---|---|---|
| `CZTCTL_OS` | Operating system (read-only) | `runtime.GOOS` |
| `CZTCTL_ARCH` | System architecture (read-only) | `runtime.GOARCH` |
| `CZTCTL_HOME` | cztctl home directory | `~/.cztctl` |
| `CZTCTL_CACHE` | Cache directory | `~/.cztctl/cache` |
| `CZTCTL_VERSION` | Current version (read-only) | Build version number |
| `CZTCTL_EXPERIMENTAL` | Experimental features toggle | `off` |

### CZTCTL_EXPERIMENTAL

Controls DSL parser selection:

- `off` (default): Use the ANTLR4 parser
- `on`: Use the handwritten recursive-descent parser

```bash
# Switch to the handwritten parser
cztctl env -w CZTCTL_EXPERIMENTAL=on

# Switch back to the ANTLR4 parser
cztctl env -w CZTCTL_EXPERIMENTAL=off
```

## Versioning

Version format: `v<go-zero-major-version>.<micro-version>`

- The first two segments (e.g. `1.10`) correspond to the go-zero version being depended upon
- The last segment is the cztctl self-incrementing version number
- `--version` output is automatically associated with the go-zero version in go.mod

## License

Same as go-zero, MIT License.
