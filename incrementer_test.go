package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestGetIncrementer(t *testing.T) {
	getNextId := GetIncrementer(0)
	res := getNextId()
	if res != 1 {
		t.Errorf("Expected next id to be 1, but received '%d'", res)
	}
}

func TestGetIncrementerWithNonZeroSeed(t *testing.T) {
	getNextId := GetIncrementer(10)
	res := getNextId()
	if res != 11 {
		t.Errorf("Expected next id to be 11, but received '%d'", res)
	}
}

func TestGetIncrementerAtomicity(t *testing.T) {
	getNextId := GetIncrementer(0)
	rand.Seed(time.Now().UnixNano())
	delays := rand.Perm(10)
	wg := &sync.WaitGroup{}
	var n uint64 = 10
	c1 := make(chan uint64, n)
	for _, delay := range delays {
		wg.Add(1)
		go func(d int) {
			defer wg.Done()
			time.Sleep(time.Duration(d) * time.Nanosecond)
			c1 <- getNextId()
		}(delay)
	}
	wg.Wait()
	close(c1)

	foundN := false
	for res := range c1 {
		if res == n {
			foundN = true
		}
	}
	if !foundN {
		t.Errorf("Did not find the max 10 id")
	}
}
