# snake

English | [中文](./readme-cn.md)

A distributed unique ID generator based on the Snowflake algorithm, designed for the [go-zero](https://github.com/zeromicro/go-zero) framework. It supports **automatic WorkerID assignment**, **clock skew tolerance**, and **concurrent safety**.

## Features

- ❄️ **Snowflake Algorithm** — 64-bit ID: 1 sign bit + timestamp + WorkerID + sequence number, IDs are monotonically increasing
- 🔧 **Flexible Bit Allocation** — `WorkerIDBits` / `SequenceBits` are configurable to accommodate clusters of different scales
- 🤖 **Automatic WorkerID Assignment** — When no WorkerID is specified, it is automatically derived from a hash of `POD_IP` or the local machine IP
- ⏱️ **Clock Skew Tolerance** — Small backward clock drifts are automatically waited out; drifts exceeding the threshold return an error
- 🔒 **Concurrent Safety** — Lock-free generation via CAS, zero duplicates under high concurrency
- 🔍 **ID Parsing** — Extract timestamp, WorkerID, and sequence number from any generated ID

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/snake
```

## Configuration

### Conf

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `WorkerIDBits` | uint8 | No | 10 | Number of bits for WorkerID, determines the maximum number of worker nodes (2^WorkerIDBits - 1) |
| `SequenceBits` | uint8 | No | 12 | Number of bits for the sequence number, determines the maximum IDs generated per millisecond (2^SequenceBits - 1) |
| `Epoch` | int64 | No | 1704067200000 | Start timestamp (milliseconds), i.e. 2024-01-01 00:00:00 UTC, used to reduce ID length |
| `TimeDifference` | int64 | No | 5 | Clock skew tolerance (milliseconds); small backward drifts within this range are automatically waited out |
| `WorkerID` | int64 | No | 0 | Manually specified WorkerID; when set to 0, it is auto-calculated from the IP address |

> **Constraint**: `WorkerIDBits + SequenceBits` must not exceed 63; otherwise `Validate()` returns an error.

## API Reference

### Constructors

| Function | Signature | Description |
|----------|-----------|-------------|
| `MustNewSnake` | `func MustNewSnake(snakeConf Conf) Snake` | Creates a Snake instance; panics if validation fails |
| `NewSnake` | `func NewSnake(conf Conf) (Snake, error)` | Creates a Snake instance; returns an error if validation fails |

> `NewSnake` internally calls `conf.Validate()` to validate the configuration, and computes `maxWorkerID`, `maxSequence`, bit shifts, and WorkerID.

### Snake Interface Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Generator` | `Generator() (int64, error)` | Generates a unique ID; concurrency-safe |
| `ParseID` | `ParseID(id int64) (timestamp int64, workerID int64, sequence int64)` | Parses the timestamp, WorkerID, and sequence number from an ID |
| `GetTimestampFromID` | `GetTimestampFromID(id int64) int64` | Extracts the timestamp (milliseconds) from an ID |
| `GetWorkerIDFromID` | `GetWorkerIDFromID(id int64) int64` | Extracts the WorkerID from an ID |
| `GetSequenceFromID` | `GetSequenceFromID(id int64) int64` | Extracts the sequence number from an ID |
| `GetTimeFromID` | `GetTimeFromID(id int64) time.Time` | Extracts the time from an ID, returning a `time.Time` object |

## Advanced Guide

### Snowflake Bit Layout

A 64-bit ID is composed of three parts (the highest bit is the sign bit, always 0):

```
| 0 | ←——— Timestamp (41 bits) ———→ | ←— WorkerID —→ | ←—— Sequence ——→ |
|   |   (currentTime - Epoch)        |  (WorkerIDBits) | (SequenceBits)   |
```

- **Timestamp**: `64 - 1 - WorkerIDBits - SequenceBits` bits, recording the difference `current millisecond timestamp - Epoch`. Under the default configuration this occupies 41 bits, providing approximately 69 years of usable range.
- **WorkerID**: Number of bits determined by `WorkerIDBits`. Default is 10 bits, supporting up to 1,024 worker nodes.
- **Sequence**: Number of bits determined by `SequenceBits`. Default is 12 bits, allowing up to 4,096 IDs per millisecond.

#### ID Assembly Formula

```go
snowflake = ((timestamp - Epoch) << timestampLeftShift) |
            (workerID << workerIDShift) |
            sequence
```

Where:
- `workerIDShift = SequenceBits`
- `timestampLeftShift = SequenceBits + WorkerIDBits`

#### Common Bit Allocation Schemes

| Scheme | WorkerIDBits | SequenceBits | Timestamp Bits | Max Nodes | IDs/ms | Usable Lifespan |
|--------|-------------|-------------|----------------|-----------|--------|-----------------|
| Default | 10 | 12 | 41 | 1,024 | 4,096 | ~69 years |
| Many Nodes | 13 | 10 | 40 | 8,192 | 1,024 | ~34 years |
| High Throughput | 8 | 14 | 41 | 256 | 16,384 | ~69 years |

### Clock Skew Handling

When the current time is detected to be earlier than the timestamp of the last generated ID, Snake adopts different strategies based on the magnitude of the drift:

| Drift Magnitude | Behavior |
|----------------|----------|
| ≤ `TimeDifference` ms | Spin-wait until the clock catches up, then generate the ID |
| > `TimeDifference` ms | Return an error immediately and refuse to generate the ID |

```go
// Clock skew tolerance is 5ms (default)
id, err := s.Generator()
if err != nil {
    // err: "clock moved backwards, refusing to generate id for X milliseconds"
}
```

> **Recommendation**: Use NTP to keep clocks synchronized in production; avoid setting `TimeDifference` too high, as it will block `Generator` for an extended period.

### Automatic WorkerID Assignment

When `Conf.WorkerID` is 0, Snake automatically computes the WorkerID with the following priority:

1. Read the `POD_IP` environment variable
2. If `POD_IP` is not set, call `netx.InternalIp()` to obtain the local intranet IP
3. Strip dots from the IP address, apply FNV-1a 32-bit hashing, then map the result modulo to the range `[0, maxWorkerID]`

```go
// In Kubernetes, inject the Pod IP via the POD_IP environment variable
// env: POD_IP=10.0.1.42

conf := snake.Conf{
    WorkerIDBits: 10,
    SequenceBits: 12,
    WorkerID:     0, // auto-assign
}
s := snake.MustNewSnake(conf)
```

> **Note**: Auto-assigned WorkerIDs are based on IP hashing; different IPs may hash to the same WorkerID (collision). When the number of nodes approaches `maxWorkerID`, it is recommended to specify WorkerID manually.

### Concurrent Safety

The `Generator` method uses CAS operations from `sync/atomic` to ensure concurrent safety:

- **Same millisecond**: Atomically increments the sequence number via `CompareAndSwapInt64`, avoiding mutex locks
- **New millisecond**: Atomically updates the timestamp via CAS and resets the sequence number to 0
- **Sequence exhausted**: Spin-waits for the next millisecond without blocking other goroutines

```go
// Concurrency-safe, no additional locking required
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        id, _ := s.Generator() // zero duplicates
        _ = id
    }()
}
wg.Wait()
```

### ID Parsing

Extract individual components from a generated ID:

```go
id, _ := s.Generator()

// Option 1: extract all components at once
timestamp, workerID, sequence := s.ParseID(id)

// Option 2: extract individual components as needed
ts  := s.GetTimestampFromID(id)  // millisecond timestamp
wid := s.GetWorkerIDFromID(id)   // WorkerID
seq := s.GetSequenceFromID(id)   // sequence number
t   := s.GetTimeFromID(id)       // time.Time object
```

> `GetTimestampFromID`, `GetWorkerIDFromID`, and `GetSequenceFromID` all delegate to `ParseID` internally.

## Full Examples

### Using with go-zero

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

    // Use the generated ID...
    timestamp, workerID, sequence := l.svcCtx.Snake.ParseID(id)
    l.Logger.Infof("generated id=%d, timestamp=%d, workerID=%d, sequence=%d",
        id, timestamp, workerID, sequence)

    return &types.CreateOrderResp{ID: id}, nil
}
```

### Standalone Usage

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
        WorkerID:       0, // auto-assign
    }

    s := snake.MustNewSnake(conf)

    // Generate an ID
    id, err := s.Generator()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Generated ID: %d\n", id)

    // Parse the ID
    timestamp, workerID, sequence := s.ParseID(id)
    fmt.Printf("Timestamp: %d, WorkerID: %d, Sequence: %d\n",
        timestamp, workerID, sequence)

    // Get a time.Time object
    t := s.GetTimeFromID(id)
    fmt.Printf("Time: %s\n", t.Format(time.RFC3339Nano))
}
```

### Manually Specifying WorkerID

```go
conf := snake.Conf{
    WorkerIDBits:   10,
    SequenceBits:   12,
    Epoch:          1704067200000,
    TimeDifference: 5,
    WorkerID:       42, // manually specified
}
s := snake.MustNewSnake(conf)
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
# Snake Snowflake ID Generator Documentation

[中文](./readme-cn.md)

## Overview

Snake is a distributed unique ID generator based on the Snowflake Algorithm, suitable for distributed systems that need to generate globally unique identifiers. It efficiently generates time-ordered unique IDs while avoiding the performance bottleneck of database auto-increment primary keys.

## Features

- **High Performance**: In-memory computation with no database dependency
- **Uniqueness**: Guarantees unique IDs in distributed environments
- **Ordering**: Generated IDs are time-ordered
- **Configurable**: Supports flexible bit allocation
- **Concurrency-Safe**: Supports ID generation in high-concurrency environments
- **Fault Tolerance**: Includes clock-backward handling mechanism

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/snake
```


## Quick Start

### 1. Create a Snake Instance

It is recommended to initialize with [MustNewSnake](./snake.go):

```go
package main

import (
    "fmt"
	
    "github.com/lerity-yao/czt-contrib/snake"
)

func main() {
    // 配置参数
    conf := snake.Conf{
        WorkerIDBits:   10,              // 工作节点ID占用位数，默认10位，最多1023个工作节点
        SequenceBits:   12,              // 序列号占用位数，默认12位，每毫秒最多生成4096个ID
        Epoch:          1704067200000,   // 起始时间戳（毫秒），默认值为2024-01-01 00:00:00
        TimeDifference: 5,               // 时钟回拨容忍度（毫秒），默认5毫秒
        WorkerID:       1,               // 工作节点ID，可选，若为0则自动根据IP计算
    }

    // 使用MustNewSnake创建实例（配置错误时会panic）
    s := snake.MustNewSnake(conf)
    
    // 生成ID
    id, err := s.Generator()
    if err != nil {
        log.Fatal("Failed to generate ID:", err)
    }
    fmt.Printf("Generated ID: %d\n", id)
}
```


### 2. Handle Initialization Errors with NewSnake

If you need to handle initialization errors, you can also use [NewSnake](./snake.go):

```go
s, err := snake.NewSnake(conf)
if err != nil {
    log.Fatal("Failed to create Snake:", err)
}
```


## Basic Usage

### Generate a Unique ID

```go
// 生成唯一ID
id, err := s.Generator()
if err != nil {
    log.Printf("Error generating ID: %v", err)
    return
}

fmt.Printf("Generated ID: %d\n", id)
```


### Parse an ID

```go
// 解析ID
timestamp, workerID, sequence := s.ParseID(id)
fmt.Printf("Timestamp: %d, WorkerID: %d, Sequence: %d\n", timestamp, workerID, sequence)

// 单独获取各部分
timestamp = s.GetTimestampFromID(id)
workerID = s.GetWorkerIDFromID(id)
sequence = s.GetSequenceFromID(id)
timeObj := s.GetTimeFromID(id) // 返回time.Time对象
```


## Configuration Details

| Field | Type | Default | Description |
|--------|------|--------|------|
| WorkerIDBits | uint8 | 10 | Number of bits used for the worker ID, determines the maximum number of workers (2^WorkerIDBits - 1) |
| SequenceBits | uint8 | 12 | Number of bits used for the sequence number, determines the maximum IDs generated per millisecond (2^SequenceBits - 1) |
| Epoch | int64 | 1704067200000 | Custom start timestamp in milliseconds, used to reduce ID length |
| TimeDifference | int64 | 5 | Clock-backward tolerance in milliseconds; maximum allowed system clock regression |
| WorkerID | int64 | 0 | Worker node ID; if 0, it is automatically calculated based on the IP address |

### Bit Allocation Example

- **Default Configuration** (WorkerIDBits=10, SequenceBits=12):
    - 41 bits total for timestamp (~69 years)
    - 10 bits for worker ID (up to 1023 nodes)
    - 12 bits for sequence number (up to 4096 IDs per millisecond)
    - 1 sign bit

### Automatic Worker ID Assignment

When `WorkerID` is set to 0, the system automatically assigns a worker ID according to the following rules:

1. Prefer reading the `POD_IP` environment variable
2. If the environment variable does not exist, obtain the local internal IP address
3. Hash the IP address to derive the worker ID

## Advanced Usage

### 1. Concurrent ID Generation

```go
func generateConcurrent(s snake.Snake, numGoroutines, idsPerGoroutine int) {
    var wg sync.WaitGroup
    idChan := make(chan int64, numGoroutines*idsPerGoroutine)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < idsPerGoroutine; j++ {
                id, err := s.Generator()
                if err != nil {
                    log.Printf("Error generating ID: %v", err)
                    continue
                }
                idChan <- id
            }
        }()
    }
    
    wg.Wait()
    close(idChan)
    
    // 收集所有ID并验证唯一性
    idSet := make(map[int64]bool)
    for id := range idChan {
        if idSet[id] {
            log.Printf("Duplicate ID detected: %d", id)
        }
        idSet[id] = true
    }
    
    fmt.Printf("Generated %d unique IDs\n", len(idSet))
}
```


### 2. ID Parsing and Validation

```go
func analyzeID(s snake.Snake, id int64) {
    timestamp, workerID, sequence := s.ParseID(id)
    
    // 打印详细信息
    t := time.UnixMilli(timestamp)
    fmt.Printf("ID Analysis:\n")
    fmt.Printf("  Full ID: %d\n", id)
    fmt.Printf("  Timestamp: %d (%s)\n", timestamp, t.Format("2006-01-02 15:04:05.000"))
    fmt.Printf("  Worker ID: %d\n", workerID)
    fmt.Printf("  Sequence: %d\n", sequence)
}
```


## Error Handling

Snake may return the following errors:

1. **Clock Backward Error**: System time regressed beyond the tolerance threshold
2. **Worker ID Error**: Manually specified WorkerID is out of range
3. **IP Retrieval Failure**: Unable to obtain an IP address when auto-assigning WorkerID

```go
id, err := s.Generator()
if err != nil {
    switch {
    case strings.Contains(err.Error(), "clock moved backwards"):
        // 处理时钟回拨错误
        log.Printf("Clock issue: %v", err)
    case strings.Contains(err.Error(), "WorkerID"):
        // 处理WorkerID错误
        log.Printf("Worker ID issue: %v", err)
    default:
        // 其他错误
        log.Printf("Other error: %v", err)
    }
}
```


## Performance Recommendations

1. **Reuse the Snake Instance**: Avoid creating Snake instances frequently; create and reuse it at program initialization
2. **Allocate Bits Reasonably**: Distribute bits among timestamp, worker node, and sequence number according to business requirements
3. **Monitor ID Generation Rate**: Monitor ID generation performance in high-concurrency scenarios
4. **Consider Time Zones**: Pay attention to the relationship between the Epoch timestamp and time zones

## Best Practices

### 1. Initialization (Recommended: MustNew)

```go
var s snake.Snake

func init() {
    conf := snake.Conf{
        WorkerIDBits:   10,
        SequenceBits:   12,
        Epoch:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
        TimeDifference: 5,
        WorkerID:       0, // 自动分配
    }
    
    // 使用MustNewSnake进行初始化，配置错误时会panic
    s = snake.MustNewSnake(conf)
}
```


### 2. Initialize at Application Startup

```go
package main

import (
    "github.com/lerity-yao/czt-contrib/snake"
)

var GlobalSnake snake.Snake

func main() {
    // 应用启动时初始化
    conf := snake.Conf{
        WorkerIDBits: 10,
        SequenceBits: 12,
        Epoch:       1704067200000,
        WorkerID:     1,
    }
    
    GlobalSnake = snake.MustNewSnake(conf)
    
    // 启动应用...
}
```


## Notes

1. **Time Synchronization**: Ensure the distributed system clocks are synchronized; NTP is recommended
2. **ID Length**: Generated IDs are 64-bit long integers; pay attention to storage and transmission compatibility
3. **Capacity Planning**: Estimate ID generation volume based on business growth and plan bit allocation reasonably
4. **Monitoring and Alerting**: Establish monitoring and alerting for ID generation failures

## FAQ

### Q: How do I choose a suitable bit configuration?

A: Decide based on the following factors:
- **WorkerIDBits**: Maximum expected number of machines
- **SequenceBits**: Expected peak QPS per machine
- **Timestamp bits**: Expected service lifetime

### Q: How is clock regression handled?

A:
- The system has a built-in clock-backward tolerance mechanism; adjust `TimeDifference` in the configuration
- Use NTP to ensure system clock synchronization
- Combine with other uniqueness constraints for critical business scenarios

### Q: How secure are the IDs?

A:
- Generated IDs are somewhat predictable because they are time-related
- If encryption is required, perform secondary processing externally
- It is not recommended to expose IDs directly as business-sensitive information

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
