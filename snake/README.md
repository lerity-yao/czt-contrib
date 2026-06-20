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
