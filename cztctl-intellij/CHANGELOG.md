# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [0.0.1] - 2026-03-22

### 新增

- 支持 `.cron` 和 `.rabbitmq` 文件的语法高亮
  - 关键字：`syntax`、`info`、`import`、`type`、`service`
  - 注解：`@server`、`@doc`、`@handler`、`@cron`、`@cronRetry`
  - 类型系统：结构体定义、字段声明、内置类型、struct tag
  - 路由行：cron 任务标识符 `TaskName(ReqType)`、rabbitmq 队列名称 `queue.name(MsgType)`
  - 注释：单行注释 `//`、块注释 `/* */`
  - 字符串：双引号字符串、反引号原始字符串
- 上下文感知的语义着色
  - info 模块：`info` 关键字加粗斜体、KV 键名独立配色
  - type 模块：`type` 关键字加粗斜体、结构体名称加粗、字段名称/类型/tag 各有独立颜色
  - @server 模块：`@server` 关键字加粗斜体、KV 键名独立配色、非字符串值斜体高亮
  - service 模块：`service` 关键字加粗斜体、服务名称/handler 名称/路由参数/注解关键字各有独立颜色
- 语法错误实时检查（红色波浪线）
- 文件类型语义校验（.rabbitmq 禁止 @cron / @cronRetry）
- 代码导航：Ctrl+Click 跳转
  - 路由参数 → 跳转到 type 定义（支持跨 import 文件解析）
  - import 路径字符串 → 打开对应文件
- 快速创建文件：右键 New → Cron File (.cron) / RabbitMQ File (.rabbitmq)
- 提供 18 个 Live Templates 代码模板
  - 顶层结构：`syntax`、`info`、`import`、`type`、`typeg`
  - @server：`server`、`serverfull`
  - service：`service`
  - 注解：`@doc`、`@dockv`、`@handler`、`@cron`、`@cronRetry`
  - 组合模板：`crontask`、`exttask`、`consumer`
  - 文件模板：`cronfile`、`mqfile`
- 兼容 GoLand 2024.3 ~ 2025.3.x
- 添加 ANTLR4 语法定义文件（`Cztctl.g4`），与 CLI 解析器共享统一语法规范
