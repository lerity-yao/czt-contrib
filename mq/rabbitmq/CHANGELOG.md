# Changelog

## v0.1.3

### 新增功能

#### 1. 连接可靠性增强

- **NotifyClose 监听**: Sender 和 Listener 同时监听 Connection 和 Channel 的关闭事件
- **自动重连**: 连接断开后自动尝试重连（最多 10 次）
- **线程安全**: 使用 `atomic.Bool` 保证 `closed` 标志线程安全
- **双重检查**: 避免重复重连（加锁前检查 + 加锁后检查）
- **优雅停机**: 收到 go-zero proc 停止信号后不再重连

#### 2. Sender 优雅停机
- 自动注册 `proc.AddShutdownListener` 关闭钩子
- 停机时标记 `closed`，不再重连

#### 3. 监控指标大幅扩展

**Sender 指标（5个）**：

| 指标名 | 类型 | Labels | 说明 |
|--------|------|--------|------|
| `mq_sender_send_total` | Counter | exchange, route_key, status | 发送总数 |
| `mq_sender_send_duration_ms` | Histogram | exchange, route_key | 发送耗时 |
| `mq_sender_send_size_bytes` | Histogram | exchange, route_key | 消息大小 |
| `mq_sender_reconnect_total` | Counter | - | 重连次数 |
| `mq_sender_disconnect_total` | Counter | - | 掉线次数 |

**Listener 指标（9个）**：

| 指标名 | 类型 | Labels | 说明 |
|--------|------|--------|------|
| `mq_listener_consume_total` | Counter | queue, status | 消费总数 |
| `mq_listener_consume_duration_ms` | Histogram | queue | 消费耗时 |
| `mq_listener_consume_size_bytes` | Histogram | queue | 消息大小 |
| `mq_listener_in_flight` | Gauge | queue | 当前处理中消息数 |
| `mq_listener_parse_error_total` | Counter | queue | 解析失败数 |
| `mq_listener_panic_total` | Counter | queue | Panic 次数 |
| `mq_listener_ack_total` | Counter | queue, type | ACK/Reject 计数 |
| `mq_listener_reconnect_total` | Counter | - | 重连次数 |
| `mq_listener_disconnect_total` | Counter | - | 掉线次数 |

#### 4. 拦截器优化
- 调整顺序为：`recovery → prometheus → logging → trace`
- 确保所有拦截器都能执行（即使 trace 解析失败）

#### 5. 其他优化
- `StartConsumerSpan` 增加 carrier nil 防护
- 消息解析移至 traceInterceptor，避免重复解析
- ACK/Reject 逻辑统一放到 processMessage 外层

### 删除的功能
- 删除 `client.go` 文件
- 删除 `parseMessage` 方法（消息解析移至 traceInterceptor）
- 删除 `requeueMessage` 重复消费逻辑

#### 1. 指标名称变更
```diff
- mq_consume_total
+ mq_listener_consume_total

- mq_consume_duration_ms
+ mq_listener_consume_duration_ms
```
**影响**: 需要更新 Prometheus 查询和 Grafana 仪表盘

#### 2. client.go 已删除
```diff
- client, err := rabbitmq.NewClient(ctx, conf)
+ sender, err := rabbitmq.NewSender(conf)
// 或
+ sender := rabbitmq.MustNewSender(ctx, conf)
```

#### 3. Sender 自动注册 shutdown hook
- `NewSender` 和 `MustNewSender` 会自动注册 `proc.AddShutdownListener`
- 在 go-zero 环境下不需要手动调用 `Close()`
- 非 go-zero 环境仍需手动调用 `Close()`

#### 4. MustNewListener 去掉 ctx 参数
```diff
- rabbitmq.MustNewListener(ctx, conf, handler)
+ rabbitmq.MustNewListener(conf, handler)
```
**说明**: 消费者内部直接使用 `context.Background()` 派生，外部传入的 ctx 无实际作用
