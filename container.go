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

func CreateHashProcessor() *HashContainer {
	c := &HashContainer{
		hashes: make(map[uint64]*HashResult),
		stats:  Stats{Total: 0, Average: 0},
		worker: &sync.WaitGroup{},
	}
	return c
}

func (c *HashContainer) StopAcceptingWork() {
	close(c.jobs)
}

func (c *HashContainer) StartAcceptingWork() {
	c.jobs = make(chan HashRequest)
	for input := range c.jobs {
		c.worker.Add(1)
		go func(req HashRequest) {
			defer c.worker.Done()
			defer c.TrackTime(time.Now())
			// This just parks the goroutine and does not consume any CPU cycles
			// https://stackoverflow.com/questions/32147421/behavior-of-sleep-and-select-in-go
			// https://xwu64.github.io/2019/02/27/Understanding-Golang-sleep-function/
			time.Sleep(5 * time.Second)
			hash, err := HashPassword(req.password)
			c.SetHashById(req.id, &HashResult{hash: hash, err: err})
		}(input)
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (c *HashContainer) DoWork(req *HashRequest) {
	if c.jobs == nil {
		panic("StartAcceptingWork needs to be called before sending work")
	}
	c.jobs <- *req
}

func (c *HashContainer) SetHashById(id uint64, hash *HashResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hashes[id] = hash
}

func (c *HashContainer) GetHashById(id uint64) (*HashResult, error) {
	value, exists := c.hashes[id]
	if exists {
		return value, nil
	}
	return nil, fmt.Errorf("ID '%d' not found", id)
}

func (c *HashContainer) TrackTime(start time.Time) {
	elapsed := time.Since(start)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.Average = ((c.stats.Average * c.stats.Total) + uint64(elapsed.Nanoseconds())) / (c.stats.Total + 1)
	c.stats.Total = c.stats.Total + 1
}
