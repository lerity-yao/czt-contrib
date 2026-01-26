package snake

import (
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/netx"
)

type (
	Snake interface {
		Generator() (int64, error)
		ParseID(id int64) (timestamp int64, workerID int64, sequence int64)
		GetTimestampFromID(id int64) int64
		GetWorkerIDFromID(id int64) int64
		GetSequenceFromID(id int64) int64
		GetTimeFromID(id int64) time.Time
	}

	CommonSnake struct {
		snakeConf          Conf
		maxWorkerID        int64
		maxSequence        int64
		workerIDShift      uint8
		timestampLeftShift uint8
		workerID           int64
		timestamp          int64 // last timestamp
		sequence           int64 // sequence number
	}
)

func MustNewSnake(snakeConf Conf) Snake {
	snake, err := NewSnake(snakeConf)
	logx.Must(err)
	return snake
}

func NewSnake(conf Conf) (Snake, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	c := &CommonSnake{snakeConf: conf}
	c.CalculateMaxWorkerID()
	c.CalculateMaxSequence()
	c.CalculateWorkerIDShift()
	c.CalculateTimestampLeftShift()
	if err := c.CalculateWorkerID(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *CommonSnake) CalculateMaxWorkerID() {
	if c.snakeConf.WorkerIDBits >= 63 {
		c.maxWorkerID = 1<<63 - 1
	} else {
		c.maxWorkerID = -1 ^ (-1 << c.snakeConf.WorkerIDBits)
	}
}

func (c *CommonSnake) CalculateMaxSequence() {
	if c.snakeConf.SequenceBits >= 63 {
		c.maxSequence = 1<<63 - 1
	} else {
		c.maxSequence = -1 ^ (-1 << c.snakeConf.SequenceBits)
	}
}

func (c *CommonSnake) CalculateWorkerIDShift() {
	c.workerIDShift = c.snakeConf.SequenceBits
}

func (c *CommonSnake) CalculateTimestampLeftShift() {
	c.timestampLeftShift = c.snakeConf.SequenceBits + c.snakeConf.WorkerIDBits
}

func (c *CommonSnake) CalculateWorkerID() error {
	if c.snakeConf.WorkerID != 0 {
		if c.snakeConf.WorkerID < 0 || c.snakeConf.WorkerID > c.maxWorkerID {
			return fmt.Errorf("WorkerID %d is out of range [0, %d]", c.snakeConf.WorkerID, c.maxWorkerID)
		}
		c.workerID = c.snakeConf.WorkerID
		return nil
	}

	ip := os.Getenv(envPodIP)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}

	if len(ip) == 0 {
		return fmt.Errorf("cannot get IP address, worker id must be not empty")
	}

	ipClean := strings.ReplaceAll(ip, ".", "")

	h := fnv.New32a()
	_, err := h.Write([]byte(ipClean))
	if err != nil {
		return fmt.Errorf("failed to hash IP address: %v", err)
	}
	hashValue := int64(h.Sum32())
	c.workerID = (hashValue & 0x7FFFFFFF) % (c.maxWorkerID + 1)

	// 确保workerID在有效范围内
	if c.workerID < 0 || c.workerID > c.maxWorkerID {
		return fmt.Errorf("calculated workerID %d is out of range [0, %d], setting to 0", c.workerID, c.maxWorkerID)
	}
	return nil
}

func (c *CommonSnake) Generator() (int64, error) {

	currentTime := time.Now().UnixMilli()

	lastTimestamp := atomic.LoadInt64(&c.timestamp)
	var newTimestamp, newSequence int64

	for {
		if currentTime < lastTimestamp {
			timeDifference := lastTimestamp - currentTime
			if timeDifference <= c.snakeConf.TimeDifference {

				waitUntil := time.Now().Add(time.Duration(c.snakeConf.TimeDifference) * time.Millisecond)
				for time.Now().Before(waitUntil) && currentTime < lastTimestamp {
					time.Sleep(100 * time.Microsecond) //
					currentTime = time.Now().UnixMilli()
				}

				if currentTime < lastTimestamp {
					return 0, fmt.Errorf("clock moved backwards, refusing to generate id for %d milliseconds", timeDifference)
				}
			} else {
				return 0, fmt.Errorf("clock moved backwards, refusing to generate id for %d milliseconds", timeDifference)
			}
		}

		if currentTime == lastTimestamp {
			currentSequence := atomic.LoadInt64(&c.sequence)

			if currentSequence >= c.maxSequence {
				waitUntil := time.Now().Add(100 * time.Millisecond)

				for currentTime <= lastTimestamp {
					if time.Now().After(waitUntil) {
						return 0, fmt.Errorf("timeout waiting for next millisecond")
					}
					time.Sleep(100 * time.Microsecond)
					currentTime = time.Now().UnixMilli()
				}

				continue
			}

			newSequence = currentSequence + 1
			if atomic.CompareAndSwapInt64(&c.sequence, currentSequence, newSequence) {
				newTimestamp = currentTime
				break
			}
		} else {
			if atomic.CompareAndSwapInt64(&c.timestamp, lastTimestamp, currentTime) {
				atomic.StoreInt64(&c.sequence, 0)
				newTimestamp = currentTime
				newSequence = 0
				break
			}
		}

		lastTimestamp = atomic.LoadInt64(&c.timestamp)
	}

	snowflake := ((newTimestamp - c.snakeConf.Epoch) << c.timestampLeftShift) |
		(c.workerID << c.workerIDShift) |
		newSequence

	return snowflake, nil
}

func (c *CommonSnake) ParseID(id int64) (timestamp int64, workerID int64, sequence int64) {
	sequence = id & c.maxSequence
	workerID = (id >> c.workerIDShift) & c.maxWorkerID
	timestamp = (id >> c.timestampLeftShift) + c.snakeConf.Epoch

	return timestamp, workerID, sequence
}

func (c *CommonSnake) GetTimestampFromID(id int64) int64 {
	timestamp, _, _ := c.ParseID(id)
	return timestamp
}

func (c *CommonSnake) GetWorkerIDFromID(id int64) int64 {
	_, workerID, _ := c.ParseID(id)
	return workerID
}

func (c *CommonSnake) GetSequenceFromID(id int64) int64 {
	_, _, sequence := c.ParseID(id)
	return sequence
}

func (c *CommonSnake) GetTimeFromID(id int64) time.Time {
	timestamp, _, _ := c.ParseID(id)
	return time.UnixMilli(timestamp)
}
