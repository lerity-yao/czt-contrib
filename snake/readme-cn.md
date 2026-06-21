# snake

[English](./README.md)

基于雪花算法（Snowflake）的分布式唯一 ID 生成器，专为 [go-zero](https://github.com/zeromicro/go-zero) 框架设计，支持 **WorkerID 自动分配**、**时钟回拨容忍**、**并发安全**。

## 特性

- ❄️ **雪花算法** — 64 位 ID：1 位符号 + 时间戳 + WorkerID + 序列号，ID 趋势递增
- 🔧 **灵活位分配** — WorkerIDBits / SequenceBits 可配置，适应不同规模集群
- 🤖 **WorkerID 自动分配** — 未指定 WorkerID 时，基于 POD_IP 或本机 IP 哈希自动计算
- ⏱️ **时钟回拨容忍** — 小幅回拨自动等待恢复，超出阈值返回错误
- 🔒 **并发安全** — 基于 CAS 无锁生成，高并发下零重复
- 🔍 **ID 反解** — 一键从 ID 中提取时间戳、WorkerID、序列号

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/snake
```

## 配置参数

### Conf

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `WorkerIDBits` | uint8 | 否 | 10 | WorkerID 占用位数，决定最大工作节点数（2^WorkerIDBits - 1） |
| `SequenceBits` | uint8 | 否 | 12 | 序列号占用位数，决定每毫秒最大 ID 生成数（2^SequenceBits - 1） |
| `Epoch` | int64 | 否 | 1704067200000 | 起始时间戳（毫秒），即 2024-01-01 00:00:00 UTC，用于减少 ID 长度 |
| `TimeDifference` | int64 | 否 | 5 | 时钟回拨容忍度（毫秒），小幅回拨在此范围内自动等待恢复 |
| `WorkerID` | int64 | 否 | 0 | 手动指定 WorkerID；为 0 时自动根据 IP 计算 |

> **约束**：`WorkerIDBits + SequenceBits` 不得超过 63，否则 `Validate()` 返回错误。

## API 参考

### 构造函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `MustNewSnake` | `func MustNewSnake(snakeConf Conf) Snake` | 创建 Snake 实例，校验失败 panic |
| `NewSnake` | `func NewSnake(conf Conf) (Snake, error)` | 创建 Snake 实例，校验失败返回 error |

> `NewSnake` 内部自动调用 `conf.Validate()` 校验配置，并计算 `maxWorkerID`、`maxSequence`、位移量与 WorkerID。

### Snake 接口方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `Generator` | `Generator() (int64, error)` | 生成唯一 ID，并发安全 |
| `ParseID` | `ParseID(id int64) (timestamp int64, workerID int64, sequence int64)` | 从 ID 中解析出时间戳、WorkerID、序列号 |
| `GetTimestampFromID` | `GetTimestampFromID(id int64) int64` | 从 ID 中提取时间戳（毫秒） |
| `GetWorkerIDFromID` | `GetWorkerIDFromID(id int64) int64` | 从 ID 中提取 WorkerID |
| `GetSequenceFromID` | `GetSequenceFromID(id int64) int64` | 从 ID 中提取序列号 |
| `GetTimeFromID` | `GetTimeFromID(id int64) time.Time` | 从 ID 中提取时间，返回 `time.Time` 对象 |

## 进阶指南

### 雪花算法位分配原理

64 位 ID 由三部分拼接而成（最高位为符号位，始终为 0）：

```
| 0 | ←——— 时间戳（41 位）———→ | ←— WorkerID —→ | ←—— 序列号 ——→ |
|   |   (currentTime - Epoch)   |  (WorkerIDBits)| (SequenceBits) |
```

- **时间戳**：`64 - 1 - WorkerIDBits - SequenceBits` 位，记录 `当前毫秒时间戳 - Epoch` 的差值。默认配置下占 41 位，约可用 69 年。
- **WorkerID**：由 `WorkerIDBits` 决定位数。默认 10 位，最多支持 1024 个工作节点。
- **序列号**：由 `SequenceBits` 决定位数。默认 12 位，每毫秒最多生成 4096 个 ID。

#### ID 组装公式

```go
snowflake = ((timestamp - Epoch) << timestampLeftShift) |
            (workerID << workerIDShift) |
            sequence
```

其中：
- `workerIDShift = SequenceBits`
- `timestampLeftShift = SequenceBits + WorkerIDBits`

#### 常见位分配方案

| 方案 | WorkerIDBits | SequenceBits | 时间戳位数 | 最大节点数 | 每毫秒 ID 数 | 可用年限 |
|------|-------------|-------------|-----------|-----------|-------------|---------|
| 默认 | 10 | 12 | 41 | 1,024 | 4,096 | ~69 年 |
| 多节点 | 13 | 10 | 40 | 8,192 | 1,024 | ~34 年 |
| 高并发 | 8 | 14 | 41 | 256 | 16,384 | ~69 年 |

### 时钟回拨处理

当检测到当前时间小于上次生成 ID 的时间戳时，Snake 会根据回拨幅度采取不同策略：

| 回拨幅度 | 处理方式 |
|---------|---------|
| ≤ `TimeDifference` 毫秒 | 自旋等待，直到时钟追上后再生成 ID |
| > `TimeDifference` 毫秒 | 直接返回错误，拒绝生成 ID |

```go
// 回拨容忍度为 5ms（默认值）
id, err := s.Generator()
if err != nil {
    // err: "clock moved backwards, refusing to generate id for X milliseconds"
}
```

> **建议**：生产环境使用 NTP 保持时钟同步；`TimeDifference` 不宜设置过大，否则会阻塞 Generator 较长时间。

### WorkerID 自动分配

当 `Conf.WorkerID` 为 0 时，Snake 按以下优先级自动计算 WorkerID：

1. 读取环境变量 `POD_IP`
2. 若 `POD_IP` 不存在，调用 `netx.InternalIp()` 获取本机内网 IP
3. 对 IP 地址去点后进行 FNV-1a 32 位哈希，再取模映射到 `[0, maxWorkerID]` 范围

```go
// Kubernetes 中通过 POD_IP 环境变量注入 Pod IP
// env: POD_IP=10.0.1.42

conf := snake.Conf{
    WorkerIDBits: 10,
    SequenceBits: 12,
    WorkerID:     0, // 自动分配
}
s := snake.MustNewSnake(conf)
```

> **注意**：自动分配的 WorkerID 基于 IP 哈希，不同 IP 可能哈希到同一 WorkerID（碰撞）。在节点数接近 `maxWorkerID` 时，建议手动指定 WorkerID。

### 并发安全

`Generator` 方法内部使用 `sync/atomic` 的 CAS 操作保证并发安全：

- **同毫秒**：通过 `CompareAndSwapInt64` 对序列号原子递增，避免互斥锁
- **新毫秒**：通过 CAS 原子更新时间戳，并将序列号重置为 0
- **序列号耗尽**：自旋等待下一毫秒，不会阻塞其他 goroutine

```go
// 并发安全，无需额外加锁
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        id, _ := s.Generator() // 零重复
        _ = id
    }()
}
wg.Wait()
```

### ID 反解

从已生成的 ID 中提取各部分信息：

```go
id, _ := s.Generator()

// 方式一：一次提取所有部分
timestamp, workerID, sequence := s.ParseID(id)

// 方式二：按需单独提取
ts  := s.GetTimestampFromID(id)  // 毫秒时间戳
wid := s.GetWorkerIDFromID(id)   // WorkerID
seq := s.GetSequenceFromID(id)   // 序列号
t   := s.GetTimeFromID(id)       // time.Time 对象
```

> `GetTimestampFromID`、`GetWorkerIDFromID`、`GetSequenceFromID` 内部均委托 `ParseID` 实现。

## 完整示例

### 在 go-zero 中使用

```go
// internal/config/config.go
type Config struct {
    rest.RestConf
    SnakeConf snake.Conf
}
```

```yaml
# etc/config.yaml
Name: order-api
Host: 0.0.0.0
Port: 8888

SnakeConf:
  WorkerIDBits: 10
  SequenceBits: 12
  Epoch: 1704067200000
  TimeDifference: 5
  WorkerID: 0
```

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config config.Config
    Snake  snake.Snake
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config: c,
        Snake:  snake.MustNewSnake(c.SnakeConf),
    }
}
```

```go
// internal/logic/createorderlogic.go
func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (*types.CreateOrderResp, error) {
    id, err := l.svcCtx.Snake.Generator()
    if err != nil {
        return nil, err
    }

    // 使用生成的 ID...
    timestamp, workerID, sequence := l.svcCtx.Snake.ParseID(id)
    l.Logger.Infof("generated id=%d, timestamp=%d, workerID=%d, sequence=%d",
        id, timestamp, workerID, sequence)

    return &types.CreateOrderResp{ID: id}, nil
}
```

### 独立使用

```go
package main

import (
    "fmt"
    "time"

    "github.com/lerity-yao/czt-contrib/snake"
)

func main() {
    conf := snake.Conf{
        WorkerIDBits:   10,
        SequenceBits:   12,
        Epoch:          1704067200000, // 2024-01-01 00:00:00 UTC
        TimeDifference: 5,
        WorkerID:       0, // 自动分配
    }

    s := snake.MustNewSnake(conf)

    // 生成 ID
    id, err := s.Generator()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Generated ID: %d\n", id)

    // 解析 ID
    timestamp, workerID, sequence := s.ParseID(id)
    fmt.Printf("Timestamp: %d, WorkerID: %d, Sequence: %d\n",
        timestamp, workerID, sequence)

    // 获取 time.Time 对象
    t := s.GetTimeFromID(id)
    fmt.Printf("Time: %s\n", t.Format(time.RFC3339Nano))
}
```

### 手动指定 WorkerID

```go
conf := snake.Conf{
    WorkerIDBits:   10,
    SequenceBits:   12,
    Epoch:          1704067200000,
    TimeDifference: 5,
    WorkerID:       42, // 手动指定
}
s := snake.MustNewSnake(conf)
```

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)
