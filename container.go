package main

import (
	"fmt"
	"sync"
	"time"
)

type Result struct {
	hash string
	err  error
}

type Stats struct {
	Total   uint64
	Average uint64
}

type HashContainer struct {
	mu     sync.Mutex
	hashes map[uint64]Result
	stats  Stats
}

func (c *HashContainer) setHashById(id uint64, hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hashes[id] = Result{hash: hash, err: nil}
}

func (c *HashContainer) getHashById(id uint64) (Result, error) {
	value, exists := c.hashes[id]
	if exists {
		return value, nil
	}
	return Result{}, fmt.Errorf("ID '%d' does not exist", id)
}

func (c *HashContainer) trackTime(start time.Time) {
	elapsed := time.Since(start)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.Average = ((c.stats.Average * c.stats.Total) + uint64(elapsed.Nanoseconds())) / (c.stats.Total + 1)
	c.stats.Total = c.stats.Total + 1
}
