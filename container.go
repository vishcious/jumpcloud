package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type HashResult struct {
	hash string
	err  error
}

type HashRequest struct {
	id       uint64
	password string
}

type Stats struct {
	Total   uint64
	Average uint64
}

type HashContainer struct {
	mu     sync.Mutex
	hashes map[uint64]*HashResult
	stats  Stats
	jobs   chan HashRequest
	worker *sync.WaitGroup
}

func createHashProcessor() *HashContainer {
	c := &HashContainer{
		hashes: make(map[uint64]*HashResult),
		stats:  Stats{Total: 0, Average: 0},
		jobs:   make(chan HashRequest),
		worker: &sync.WaitGroup{},
	}
	return c
}

func (c *HashContainer) shutdown() {
	close(c.jobs)
}

func (c *HashContainer) doWork() {
	for input := range c.jobs {
		c.worker.Add(1)
		go func(req HashRequest) {
			defer c.worker.Done()
			defer c.trackTime(time.Now())
			// This just parks the goroutine and does not consume any CPU cycles
			// https://stackoverflow.com/questions/32147421/behavior-of-sleep-and-select-in-go
			// https://xwu64.github.io/2019/02/27/Understanding-Golang-sleep-function/
			time.Sleep(5 * time.Second)
			hash, err := hashPassword(req.password)
			c.setHashById(req.id, &HashResult{hash: hash, err: err})
		}(input)
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (c *HashContainer) setHashById(id uint64, hash *HashResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hashes[id] = hash
}

func (c *HashContainer) getHashById(id uint64) (*HashResult, error) {
	value, exists := c.hashes[id]
	if exists {
		return value, nil
	}
	return nil, fmt.Errorf("ID '%d' not found", id)
}

func (c *HashContainer) trackTime(start time.Time) {
	elapsed := time.Since(start)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.Average = ((c.stats.Average * c.stats.Total) + uint64(elapsed.Nanoseconds())) / (c.stats.Total + 1)
	c.stats.Total = c.stats.Total + 1
}
