package id

import (
	"errors"
	"sync"
	"time"
)

// SnowflakeID is a minimal snowflake-like generator:
// 41 bits timestamp (ms) | 10 bits worker | 12 bits sequence.
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

func NewSnowflake(workerID int64) (*SnowflakeID, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, errors.New("worker_id out of range (0~1023)")
	}
	return &SnowflakeID{
		epochMS:  1700000000000, // 2023-11-14T22:13:20Z (fixed epoch)
		workerID: workerID,
	}, nil
}

func (g *SnowflakeID) Next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	nowMS := time.Now().UnixMilli()
	if nowMS < g.lastMS {
		// clock moved backwards; wait until lastMS
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
