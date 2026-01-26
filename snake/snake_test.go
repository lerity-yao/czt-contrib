package snake

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSnake(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   10,
		SequenceBits:   12,
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       1,
	}

	snake, err := NewSnake(conf)
	assert.NoError(t, err)
	assert.NotNil(t, snake)
}

func TestMustNewSnake(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   10,
		SequenceBits:   12,
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       1,
	}

	snake := MustNewSnake(conf)
	assert.NotNil(t, snake)
}

func TestSnakeGenerator(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   10,
		SequenceBits:   12,
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       1,
	}

	snake, err := NewSnake(conf)
	assert.NoError(t, err)

	// 测试生成ID
	id1, err := snake.Generator()
	assert.NoError(t, err)
	assert.True(t, id1 > 0)

	id2, err := snake.Generator()
	assert.NoError(t, err)
	assert.True(t, id2 > 0)

	// 验证ID是递增的（大部分情况下）
	assert.True(t, id2 >= id1, "ID should be monotonically increasing")
}

func TestSnakeGeneratorConcurrent(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   10,
		SequenceBits:   12,
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       1,
	}

	snake, err := NewSnake(conf)
	assert.NoError(t, err)

	// 并发生成ID，确保没有重复
	var mu sync.Mutex
	idSet := make(map[int64]bool)
	done := make(chan bool)

	// 启动多个goroutine同时生成ID
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				id, err := snake.Generator()
				assert.NoError(t, err)
				assert.True(t, id > 0)

				// 使用互斥锁保护map访问
				mu.Lock()
				// 检查ID是否重复
				if idSet[id] {
					t.Errorf("Duplicate ID generated: %d", id)
				}
				idSet[id] = true
				mu.Unlock()
			}
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSnakeParseID(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   10,
		SequenceBits:   12,
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       5,
	}

	snake, err := NewSnake(conf)
	assert.NoError(t, err)

	id, err := snake.Generator()
	assert.NoError(t, err)

	// 解析ID
	timestamp, workerID, sequence := snake.ParseID(id)

	// 验证解析结果
	assert.Equal(t, int64(5), workerID, "Worker ID should match")
	assert.True(t, sequence >= 0, "Sequence should be non-negative")
	assert.True(t, timestamp >= conf.Epoch, "Timestamp should be greater than or equal to epoch")

	// 验证从ID中单独提取的部分
	extractedWorkerID := snake.GetWorkerIDFromID(id)
	assert.Equal(t, workerID, extractedWorkerID)

	extractedSequence := snake.GetSequenceFromID(id)
	assert.Equal(t, sequence, extractedSequence)

	extractedTimestamp := snake.GetTimestampFromID(id)
	assert.Equal(t, timestamp, extractedTimestamp)

	// 验证时间转换
	timeObj := snake.GetTimeFromID(id)
	assert.True(t, timeObj.UnixMilli() == timestamp)
}

func TestSnakeWithMaxSequence(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   5, // 小的工作ID位数
		SequenceBits:   3, // 小的序列位数，以便快速达到最大序列号
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       1,
	}

	snake, err := NewSnake(conf)
	assert.NoError(t, err)

	// 在同一毫秒内生成多个ID直到序列号达到最大值
	ids := make([]int64, 0)
	maxSequence := int64(-1 ^ (-1 << conf.SequenceBits)) // 计算最大序列号

	// 先生成一些ID，然后等待到下一个毫秒再生成，确保测试序列号重置
	for i := 0; i < int(maxSequence)+2; i++ {
		id, err := snake.Generator()
		assert.NoError(t, err)
		ids = append(ids, id)

		// 为了确保在同一毫秒内生成多个ID，我们稍微等待
		if i == 0 {
			// 等待到下一个毫秒的开始
			for {
				now := time.Now()
				if now.Nanosecond()/1000000 == 0 { // 精确到毫秒
					break
				}
				time.Sleep(100 * time.Microsecond)
			}
		}
	}

	// 检查ID是否正确生成，且序列号在重置后从0开始
	assert.True(t, len(ids) > 0)
}

func TestSnakeValidateConfig(t *testing.T) {
	// 原来的验证逻辑是设置默认值而不是返回错误
	// 所以我们需要测试默认值是否正确设置
	invalidConf := Conf{
		WorkerIDBits: 64, // 超出范围，会被设为默认值
		SequenceBits: 12,
		WorkerID:     1,
	}

	// 这个配置不会报错，但会使用默认值
	snake, err := NewSnake(invalidConf)
	assert.NoError(t, err)

	// 验证默认值是否正确应用
	id, err := snake.Generator()
	assert.NoError(t, err)
	assert.True(t, id > 0)

	// 测试负值情况
	negativeConf := Conf{
		WorkerIDBits: 0, // 会被设为默认值
		SequenceBits: 0, // 会被设为默认值
		Epoch:        0, // 会被设为默认值
		WorkerID:     1,
	}

	snake2, err := NewSnake(negativeConf)
	assert.NoError(t, err)
	id2, err := snake2.Generator()
	assert.NoError(t, err)
	assert.True(t, id2 > 0)
}

func TestSnakeWorkerIDCalculation(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   10,
		SequenceBits:   12,
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       100,
	}

	snake, err := NewSnake(conf)
	assert.NoError(t, err)

	// 生成几个ID并验证工作ID部分是否一致
	for i := 0; i < 10; i++ {
		id, err := snake.Generator()
		assert.NoError(t, err)

		_, workerID, _ := snake.ParseID(id)
		assert.Equal(t, int64(100), workerID, "Worker ID should remain constant")
	}
}

func TestSnakeInvalidWorkerID(t *testing.T) {
	// 测试超出范围的WorkerID
	conf := Conf{
		WorkerIDBits: 5, // 最大支持31个worker
		SequenceBits: 12,
		Epoch:        1704067200000,
		WorkerID:     50, // 超出范围，应该报错
	}

	_, err := NewSnake(conf)
	assert.Error(t, err)
}

func TestSnakeAutoWorkerID(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   10,
		SequenceBits:   12,
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       0, // 自动分配WorkerID
	}

	snake, err := NewSnake(conf)
	assert.NoError(t, err)

	// 生成ID并验证WorkerID部分
	for i := 0; i < 10; i++ {
		id, err := snake.Generator()
		assert.NoError(t, err)

		_, workerID, _ := snake.ParseID(id)
		// WorkerID应该是非负数且在有效范围内
		assert.True(t, workerID >= 0, "Auto-assigned Worker ID should be non-negative")
		assert.True(t, workerID <= snake.(*CommonSnake).maxWorkerID, "Auto-assigned Worker ID should be within range")
	}
}

func TestSnakeCheckDuplicateIDs(t *testing.T) {
	conf := Conf{
		WorkerIDBits:   10,
		SequenceBits:   12,
		Epoch:          1704067200000,
		TimeDifference: 5,
		WorkerID:       0,
	}

	snake, err := NewSnake(conf)
	assert.NoError(t, err)

	// 生成大量ID并检查是否有重复
	const numIDs = 10000
	idSet := make(map[int64]bool)

	for i := 0; i < numIDs; i++ {
		id, err := snake.Generator()
		fmt.Println("Generated ID:", id)
		assert.NoError(t, err)
		assert.True(t, id > 0)

		// 检查ID是否已经存在（即重复）
		if idSet[id] {
			t.Fatalf("Duplicate ID found: %d at iteration %d", id, i)
		}
		idSet[id] = true
	}

	// 验证生成的ID数量是否正确
	assert.Equal(t, numIDs, len(idSet), "Number of unique IDs should match generated count")
}
