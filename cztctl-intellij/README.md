# cztctl-intellij

GoLand 插件，[go-zero](https://go-zero.dev) 微服务框架扩展代码生成工具 [cztctl](https://github.com/lerity-yao/czt-contrib/cztctl) 的 IDE 插件。支持 `.cron`（定时任务）和 `.rabbitmq`（消息队列）DSL 文件的语法高亮、语义着色、语法错误检查、代码导航与代码模板。

## 功能

### 语法高亮 + 语义着色

基于 ANTLR4 Lexer + ExternalAnnotator 驱动，覆盖 DSL 全部语法元素，并提供上下文感知的语义配色：

| 元素       | 说明                                                                    |
|----------|-----------------------------------------------------------------------|
| 关键字      | `syntax`、`info`、`import`、`type`、`service`、`map`、`struct`（模块关键字加粗斜体） |
| 注解       | `@server`（独立配色）、`@doc`、`@handler`、`@cron`、`@cronRetry`                      |
| 类型系统     | 结构体名称（加粗）、字段名称、字段类型、struct tag 各有独立颜色                      |
| 路由行      | cron 任务标识符、rabbitmq 队列名称、路由参数（加粗斜体） |
| info / @server | KV 键名独立配色，@server 非字符串值斜体高亮                                    |
| service  | 服务名称、handler 名称、注解关键字各有独立颜色                                 |
| 注释       | 单行注释 `//`、块注释 `/* */`                                                 |
| 字符串      | 双引号字符串、反引号原始字符串                                                       |

### 语法错误检查

基于 ANTLR4 Parser 全量解析 + ExternalAnnotator，实时检测语法错误并以红色波浪线标注：

- 缺少必填声明（syntax / service）
- 括号/大括号不匹配
- 注解格式错误（@cron / @handler / @doc）
- 结构体字段定义错误
- service route 格式错误
- **文件类型语义校验**：`.rabbitmq` 文件中使用 `@cron` / `@cronRetry` 会报红线

### 代码导航

**Ctrl+Click**（或 Cmd+Click）跳转：

| 元素 | 行为 |
|------|------|
| 路由参数（如 `OrderCancelReq`） | 跳转到对应的 type 定义（支持跨 import 文件解析） |
| import 路径字符串 | 打开对应的引入文件 |

### 快速创建文件

右键目录 → **New** 菜单：

| 菜单项 | 说明 |
|--------|------|
| **Cron File (.cron)** | 创建带 `syntax = "v1"` + `info()` 基础结构的 .cron 文件 |
| **RabbitMQ File (.rabbitmq)** | 创建带 `syntax = "v1"` + `info()` 基础结构的 .rabbitmq 文件 |

### Live Templates（代码模板）

输入缩写后按 **Tab** 键即可快速展开，共 18 个模板：

| 缩写           | 说明                                           |
|--------------|----------------------------------------------|
| `syntax`     | syntax 版本声明                                   |
| `info`       | info 元信息块                                     |
| `import`     | import 单文件导入                                  |
| `type`       | type 结构体声明                                    |
| `typeg`      | type 分组结构体声明                                  |
| `server`     | @server 注解块                                   |
| `serverfull` | @server 完整注解块（tags/summary/group/middleware） |
| `service`    | service 服务块                                   |
| `@doc`       | @doc 字符串文档注解                                  |
| `@dockv`     | @doc KV 文档注解                                  |
| `@handler`   | @handler 处理器名称                                |
| `@cron`      | @cron 定时表达式                                   |
| `@cronRetry` | @cronRetry 重试次数                               |
| `crontask`   | 完整的内部定时任务（@doc + @cron + @cronRetry + @handler + 路由） |
| `exttask`    | 外部触发任务（无 @cron，仅 @doc + @handler + 路由）        |
| `consumer`   | RabbitMQ 消费者（@doc + @handler + 队列路由）          |
| `cronfile`   | .cron 完整文件模板                                  |
| `mqfile`     | .rabbitmq 完整文件模板                              |

## 安装

### 从本地安装

1. GoLand → **Settings** → **Plugins**
2. 点击齿轮图标 → **Install Plugin from Disk...**
3. 选择 `build/distributions/cztctl-intellij-<version>.zip`
4. 重启 GoLand

### 更新插件

1. GoLand → **Settings** → **Plugins** → **Installed** 标签页
2. 找到 **cztctl**，点击卸载
3. **重启** GoLand
4. 按上述「从本地安装」步骤安装新版本 zip
5. **重启** GoLand

> 也可以不卸载，直接用「Install Plugin from Disk...」选择新版本 zip 覆盖安装，然后重启。

## 构建

### 前置条件

- **GoLand** 已安装（用作编译依赖的 SDK，以及提供 JBR 作为 Java 运行环境）
- 不需要额外安装 Java 或 Gradle（使用 GoLand 自带的 JBR 和项目内的 Gradle Wrapper）

### 构建步骤

```bash
cd czt-contrib/cztctl-intellij

# JAVA_HOME 指向 GoLand 自带的 JBR
JAVA_HOME=/path/to/goland/jbr ./gradlew buildPlugin
```

构建产物位于：

```
build/distributions/cztctl-intellij-<version>.zip
```

### 修改版本号

编辑 `build.gradle.kts` 中的 `version`：

```kotlin
version = "0.0.1"  // 修改为新版本号
```

同时更新 `src/main/resources/META-INF/plugin.xml` 中的 `<change-notes>` 添加新版本说明。

### 清理重建

```bash
JAVA_HOME=/path/to/goland/jbr ./gradlew clean buildPlugin
```

## 项目结构

```
cztctl-intellij/
├── build.gradle.kts                          # Gradle 构建配置（含 ANTLR4 插件）
├── settings.gradle.kts                       # Gradle 项目设置
├── gradlew                                   # Gradle Wrapper 脚本
├── gradle/wrapper/
│   └── gradle-wrapper.properties             # Gradle 版本配置
└── src/main/
    ├── antlr/com/cztctl/intellij/parser/
    │   └── Cztctl.g4                         # ANTLR4 Combined Grammar（词法+语法）
    ├── kotlin/com/cztctl/intellij/
    │   ├── CztctlLanguage.kt                 # Language 单例
    │   ├── CztctlFileType.kt                 # FileType（.cron / .rabbitmq）
    │   ├── CztctlFile.kt                     # PsiFile 实现
    │   ├── parser/
    │   │   ├── CztctlElementTypes.kt         # ANTLR4 Token → IElementType 映射
    │   │   ├── CztctlLexerAdapter.kt         # ANTLR4 Lexer → IntelliJ Lexer 适配器
    │   │   └── CztctlParserDefinition.kt     # ParserDefinition（轻量 PSI 树）
    │   ├── highlight/
    │   │   ├── CztctlSyntaxHighlighter.kt    # Token → 颜色映射 + 语义色彩常量
    │   │   └── CztctlSyntaxHighlighterFactory.kt
    │   ├── annotator/
    │   │   └── CztctlExternalAnnotator.kt    # ANTLR4 全量解析 → 红线报错 + 语义校验 + 语义着色
    │   ├── navigation/
    │   │   └── CztctlGotoDeclarationHandler.kt  # Ctrl+Click 代码导航
    │   └── action/
    │       ├── CztctlNewCronFileAction.kt    # New → Cron File 菜单动作
    │       └── CztctlNewRabbitmqFileAction.kt # New → RabbitMQ File 菜单动作
    ├── resources/
    │   ├── META-INF/
    │   │   └── plugin.xml                    # 插件描述文件
    │   ├── liveTemplates/
    │   │   └── cztctl.xml                    # Live Templates 定义
    │   └── fileTemplates/internal/
    │       ├── Cztctl Cron File.cron.ft       # .cron 文件模板
    │       └── Cztctl RabbitMQ File.rabbitmq.ft # .rabbitmq 文件模板
    └── textmate/
        └── cztctl-syntax.tmBundle/           # TextMate 语法包（保留备用）
            ├── info.plist
            └── Syntaxes/
                └── cztctl.tmLanguage.json
```

### 关键文件说明

| 文件 | 修改场景 |
|------|--------|
| `Cztctl.g4` | 新增或修改语法规则（如新增 DSL 关键字、注解） |
| `CztctlExternalAnnotator.kt` | 新增语义校验规则、新增语义着色规则 |
| `CztctlSyntaxHighlighter.kt` | 新增 Token 类型的颜色映射、修改语义色彩常量 |
| `CztctlGotoDeclarationHandler.kt` | 新增或修改代码导航规则 |
| `cztctl.xml` | 新增或修改 Live Templates 代码模板 |
| `fileTemplates/internal/*.ft` | 修改新建文件的默认模板内容 |
| `plugin.xml` | 修改插件描述、版本说明、扩展点注册 |
| `build.gradle.kts` | 修改版本号、构建配置、GoLand 兼容版本范围 |

### GoLand 版本兼容

`build.gradle.kts` 中配置了兼容范围：

```kotlin
patchPluginXml {
    sinceBuild.set("243")    // GoLand 2024.3+
    untilBuild.set("253.*")  // 到 GoLand 2025.3.x
}
```

如需支持更新版本的 GoLand，修改 `untilBuild` 即可。

## 技术架构

```
                    Cztctl.g4
                        │
          ┌─────────────┼─────────────┐
          ▼             ▼             ▼
    CztctlLexer    CztctlParser   CztctlBaseListener
    (generated)    (generated)    (generated)
          │             │             │
          ▼             │             ▼
  CztctlLexerAdapter    │    CztctlExternalAnnotator
          │             │    (语法错误 + 语义校验)
          ▼             │             │
  CztctlSyntaxHighlighter             │
  (Token → 颜色)        │             │
          │             │             ▼
          ▼             ▼         红线报错
      语法高亮        (PSI Tree)
```

| 层 | 组件 | 作用 |
|---|---|---|
| ANTLR4 生成层 | `CztctlLexer` / `CztctlParser` | 从 `Cztctl.g4` 自动生成的词法/语法分析器 |
| Lexer 适配层 | `CztctlLexerAdapter` | 将 ANTLR4 Token 转为 IntelliJ IElementType |
| 高亮层 | `CztctlSyntaxHighlighter` | 将 IElementType 映射到 IDE 颜色方案 + 语义色彩常量 |
| PSI 层 | `CztctlParserDefinition` | 轻量 PSI 树（flat），用于 IDE 基础设施 |
| 校验 + 着色层 | `CztctlExternalAnnotator` | 后台线程运行 ANTLR4 全量解析，收集语法/语义错误 + 语义着色 |
| 导航层 | `CztctlGotoDeclarationHandler` | Ctrl+Click 路由参数跳转、import 路径跳转 |
| 动作层 | `CztctlNewCronFileAction` / `CztctlNewRabbitmqFileAction` | New 菜单快速创建文件 |

## 与 cztctl CLI 的关系

本插件与 `czt-contrib/cztctl` 命令行工具共享同一套语法设计：

| 模块 | 用途 | 语法实现 |
|------|------|----------|
| `czt-contrib/cztctl` | CLI 代码生成工具 | Go 手写解析器 |
| `czt-contrib/cztctl-intellij` | GoLand IDE 插件 | ANTLR4 Combined Grammar |

两端语法规范保持一致，权威来源为 cztctl CLI 的手写解析器。插件的 `Cztctl.g4` 需与 CLI 解析器行为对齐。

## DSL 语法示例

### .cron 文件

```
syntax = "v1"

info (
    title: "定时任务服务"
    desc: "订单超时自动取消"
)

type OrderCancelReq {
    OrderId int64 `json:"orderId"`
}

@server (
    group: order
)
service order-cron {
    @doc "订单超时取消"
    @cron "*/5 * * * *"
    @cronRetry 3
    @handler OrderCancel
    OrderCancelTask(OrderCancelReq)
}
```

### .rabbitmq 文件

```
syntax = "v1"

type PaymentMsg {
    OrderId int64  `json:"orderId"`
    Amount  string `json:"amount"`
}

service payment-mq {
    @doc "支付成功回调"
    @handler PaymentSuccess
    payment.success(PaymentMsg)
}
```
