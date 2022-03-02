package main

import (
	"sync/atomic"
)

func getIncrementer(seed uint64) func() uint64 {
	var start uint64 = seed
	return func() uint64 {
		atomic.AddUint64(&start, 1)
		return start
	}
}
