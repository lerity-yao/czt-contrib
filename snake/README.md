# Snake 雪花算法 ID 生成器使用文档

## 概述

Snake 是一个基于雪花算法（Snowflake Algorithm）的分布式唯一 ID 生成器，适用于需要生成全局唯一标识符的分布式系统场景。它能够高效地生成具有时间顺序性的唯一 ID，避免了数据库自增主键的性能瓶颈。

## 特性

- **高性能**：基于内存计算，无需数据库依赖
- **唯一性**：保证分布式环境下的 ID 唯一性
- **有序性**：生成的 ID 具有时间顺序性
- **可配置**：支持灵活配置位数分配
- **并发安全**：支持高并发环境下的 ID 生成
- **容错性**：具备时钟回拨处理机制

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/snake
```


## 快速开始

### 1. 创建 Snake 实例

推荐使用 [MustNewSnake](file:///home/yaox/code/bk/czt-contrib/snake/snake.go#L33-L37) 进行初始化：

```go
package main

import (
    "fmt"
	
    "github.com/your-repo/czt-contrib/snake"
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
    id := s.Generator()
    fmt.Printf("Generated ID: %d\n", id)
}
```


### 2. 使用 NewSnake 进行错误处理

如果你需要处理初始化错误，也可以使用 [NewSnake](file:///home/yaox/code/bk/czt-contrib/snake/snake.go#L39-L54)：

```go
s, err := snake.NewSnake(conf)
if err != nil {
    log.Fatal("Failed to create Snake:", err)
}
```


## 基本使用

### 生成唯一ID

```go
// 生成唯一ID
id, err := s.Generator()
if err != nil {
    log.Printf("Error generating ID: %v", err)
    return
}

fmt.Printf("Generated ID: %d\n", id)
```


### 解析ID

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


## 配置详解

| 配置项 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| WorkerIDBits | uint8 | 10 | 工作节点ID占用的位数，决定了最大工作节点数（2^WorkerIDBits - 1） |
| SequenceBits | uint8 | 12 | 序列号占用的位数，决定了每毫秒最大ID生成数（2^SequenceBits - 1） |
| Epoch | int64 | 1704067200000 | 自定义起始时间戳（毫秒），用于减少ID长度 |
| TimeDifference | int64 | 5 | 时钟回拨容忍度（毫秒），系统允许的最大时钟回拨时间 |
| WorkerID | int64 | 0 | 工作节点ID，若为0则自动根据IP地址计算

### 位数分配示例

- **默认配置** (WorkerIDBits=10, SequenceBits=12):
    - 总共41位用于时间戳（约69年）
    - 10位用于工作节点ID（最多1023个节点）
    - 12位用于序列号（每毫秒最多4096个ID）
    - 1位符号位

### 工作节点ID自动分配

当 `WorkerID` 设置为 0 时，系统会根据以下规则自动分配工作节点ID：

1. 优先读取环境变量 `POD_IP`
2. 如果环境变量不存在，则获取本机内部IP地址
3. 对IP地址进行哈希计算，得出工作节点ID

## 高级用法

### 1. 并发安全的ID生成

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


### 2. ID解析与验证

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


## 错误处理

Snake 可能返回以下错误：

1. **时钟回拨错误**：系统时间发生回拨超过容忍度
2. **工作节点ID错误**：手动指定的WorkerID超出范围
3. **IP获取失败**：自动分配WorkerID时无法获取IP地址

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


## 性能建议

1. **复用Snake实例**：避免频繁创建Snake实例，应在程序初始化时创建并复用
2. **合理配置位数**：根据业务需求合理分配时间戳、工作节点和序列号的位数
3. **监控ID生成速率**：在高并发场景下监控ID生成性能
4. **时区考虑**：注意Epoch时间戳与时区的关系

## 最佳实践

### 1. 初始化（推荐使用MustNew）

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


### 2. 在应用启动时初始化

```go
package main

import (
    "github.com/your-repo/czt-contrib/snake"
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
    
    GlobalSnake := snake.MustNewSnake(conf)
    
    // 启动应用...
}
```


## 注意事项

1. **时间同步**：确保分布式系统的时钟同步，推荐使用NTP服务
2. **ID长度**：生成的ID为64位长整型，注意存储和传输的兼容性
3. **容量规划**：根据业务增长预估ID生成量，合理规划位数分配
4. **监控告警**：对ID生成失败的情况建立监控和告警机制

## 常见问题

### Q: 如何选择合适的位数配置？

A: 根据以下因素决定：
- **WorkerIDBits**: 预计的最大机器数量
- **SequenceBits**: 预计的单机QPS峰值
- **时间戳位数**: 服务预计运行年限

### Q: 时钟回拨如何处理？

A:
- 系统内置时钟回拨容忍机制，可在配置中调整 `TimeDifference`
- 建议使用NTP确保系统时钟同步
- 在关键业务中可结合其他唯一性约束

### Q: ID的安全性如何？

A:
- 生成的ID具有一定可预测性（时间相关）
- 如需加密ID，可在外部进行二次处理
- 不建议直接暴露ID作为业务敏感信息
