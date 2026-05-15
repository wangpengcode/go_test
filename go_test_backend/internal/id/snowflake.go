package id

import (
	"errors"
	"sync"
	"time"
)

// SnowflakeID 是一个简化版的“雪花算法”ID 生成器：
// 41 位时间戳（毫秒）| 10 位 worker | 12 位序列号。
type SnowflakeID struct {
	mu       sync.Mutex
	epochMS  int64
	workerID int64
	lastMS   int64
	sequence int64
}

const (
	workerBits   = 10
	sequenceBits = 12

	maxWorkerID = (1 << workerBits) - 1
	maxSequence = (1 << sequenceBits) - 1
)

// NewSnowflake 创建一个 SnowflakeID 生成器。
//
// 给刚接触 Go 的同学：
// - `&SnowflakeID{...}` 表示“取地址”，也就是返回结构体指针。
// - 这里返回 error 的目的：如果 workerID 不合法，就不要创建一个“坏的”生成器。
func NewSnowflake(workerID int64) (*SnowflakeID, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, errors.New("worker_id out of range (0~1023)")
	}
	return &SnowflakeID{
		epochMS:  1700000000000, // 2023-11-14T22:13:20Z (fixed epoch)
		workerID: workerID,
	}, nil
}

// Next 生成下一个唯一 ID。
//
// 给刚接触 Go 的同学：
// - `g.mu.Lock()` / `defer g.mu.Unlock()` 用来保护并发下的共享状态（避免多个 goroutine 同时改数据）。
// - `defer` 会在函数结束时执行（哪怕中间 return 了也会执行）。
func (g *SnowflakeID) Next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	nowMS := time.Now().UnixMilli()
	if nowMS < g.lastMS {
		// 系统时间回拨了：这里简单处理为“不让时间变小”，避免生成重复 ID。
		nowMS = g.lastMS
	}

	if nowMS == g.lastMS {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for nowMS <= g.lastMS {
				nowMS = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastMS = nowMS

	tsPart := (nowMS - g.epochMS) << (workerBits + sequenceBits)
	workerPart := g.workerID << sequenceBits
	return tsPart | workerPart | g.sequence
}
