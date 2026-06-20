# Changelog

[中文](./changelog-cn.md)

All version change logs. Format based on [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/).

## [0.1.5] - 2026-06-04

### Dependencies

- `github.com/rabbitmq/amqp091-go` v1.10.0 → v1.11.0
- `github.com/zeromicro/go-zero` v1.10.0 → v1.10.2
- `go.opentelemetry.io/otel*` v1.24.0 → v1.40.0 (intentionally not upgraded to v1.44+ to avoid forcing go 1.25)
- Synced `go mod tidy` to clean up unused indirect dependencies
- `go` directive remains 1.24.0

## [0.1.4] - 2026-03-20

### Breaking Changes

#### 1. MustNewSender Removed the ctx Parameter
```diff
- rabbitmq.MustNewSender(ctx, conf)
+ rabbitmq.MustNewSender(conf)
```
**Description**: Consistent with MustNewListener; ctx is not required during initialization

## [0.1.3] - 2026-03-20

### New Features

#### 1. Connection Reliability Enhancements

- **NotifyClose Listener**: Both Sender and Listener listen for Connection and Channel close events
- **Auto Reconnect**: Automatically attempts to reconnect after disconnection (up to 10 times)
- **Thread Safety**: Uses `atomic.Bool` to ensure thread safety of the `closed` flag
- **Double-Checked Locking**: Avoids duplicate reconnections (pre-lock check + post-lock check)
- **Graceful Shutdown**: Stops reconnecting after receiving the go-zero proc stop signal

#### 2. Sender Graceful Shutdown
- Automatically registers `proc.AddShutdownListener` shutdown hook
- Marks `closed` on shutdown, preventing further reconnections

#### 3. Greatly Expanded Metrics

**Sender Metrics (5)**:

| Metric Name | Type | Labels | Description |
|--------|------|--------|------|
| `rabbitmq_sender_send_total` | Counter | exchange, route_key, status | Total sends |
| `rabbitmq_sender_send_duration_ms` | Histogram | exchange, route_key | Send duration |
| `rabbitmq_sender_send_size_bytes` | Histogram | exchange, route_key | Message size |
| `rabbitmq_sender_reconnect_total` | Counter | - | Number of reconnections |
| `rabbitmq_sender_disconnect_total` | Counter | - | Number of disconnections |

**Listener Metrics (9)**:

| Metric Name | Type | Labels | Description |
|--------|------|--------|------|
| `rabbitmq_listener_consume_total` | Counter | queue, status | Total consumed |
| `rabbitmq_listener_consume_duration_ms` | Histogram | queue | Consumption duration |
| `rabbitmq_listener_consume_size_bytes` | Histogram | queue | Message size |
| `rabbitmq_listener_in_flight` | Gauge | queue | Messages currently being processed |
| `rabbitmq_listener_parse_error_total` | Counter | queue | Number of parse failures |
| `rabbitmq_listener_panic_total` | Counter | queue | Number of panics |
| `rabbitmq_listener_ack_total` | Counter | queue, type | ACK/Reject count |
| `rabbitmq_listener_reconnect_total` | Counter | - | Number of reconnections |
| `rabbitmq_listener_disconnect_total` | Counter | - | Number of disconnections |

#### 4. Interceptor Optimization
- Order adjusted to: `recovery → prometheus → logging → trace`
- Ensures all interceptors execute even if trace parsing fails

#### 5. Other Improvements
- `StartConsumerSpan` adds carrier nil protection
- Message parsing moved to traceInterceptor to avoid duplicate parsing
- ACK/Reject logic unified to the outer processMessage layer

### Removed
- Removed `client.go`
- Removed `parseMessage` method (message parsing moved to traceInterceptor)
- Removed `requeueMessage` duplicate consumption logic

#### 1. Metric Name Changes
```diff
- mq_consume_total
+ rabbitmq_listener_consume_total

- mq_consume_duration_ms
+ rabbitmq_listener_consume_duration_ms
```
**Impact**: Prometheus queries and Grafana dashboards need to be updated

#### 2. client.go Removed
```diff
- client, err := rabbitmq.NewClient(ctx, conf)
+ sender, err := rabbitmq.NewSender(conf)
// or
+ sender := rabbitmq.MustNewSender(conf)
```

#### 3. Sender Auto-Registers Shutdown Hook
- `NewSender` and `MustNewSender` automatically register `proc.AddShutdownListener`
- In the go-zero environment, there is no need to call `Close()` manually
- Outside the go-zero environment, `Close()` still needs to be called manually

#### 4. MustNewListener Removed the ctx Parameter
```diff
- rabbitmq.MustNewListener(ctx, conf, handler)
+ rabbitmq.MustNewListener(conf, handler)
```
**Description**: The consumer internally derives from `context.Background()`; the externally passed ctx has no practical effect
