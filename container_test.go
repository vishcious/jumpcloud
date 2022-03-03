package main

import (
	"testing"
	"time"
)

func TestSetHashById(t *testing.T) {
	target := CreateHashProcessor()
	target.SetHashById(uint64(1), &HashResult{hash: "test", err: nil})
	if len(target.hashes) != 1 {
		t.Errorf("Expected one hash entry in map")
	}
	if target.hashes[uint64(1)].hash != "test" {
		t.Errorf("Expected 'hash' test result")
	}
}

func TestGetHashById(t *testing.T) {
	target := CreateHashProcessor()
	target.SetHashById(uint64(1), &HashResult{hash: "test", err: nil})
	res, _ := target.GetHashById(uint64(1))
	if res == nil {
		t.Errorf("Valid result expected")
	}
	if res != nil && res.hash != "test" {
		t.Errorf("Expected 'hash' test result")
	}
}

func TestGetHashByIdReturnsErrorForMissingId(t *testing.T) {
	target := CreateHashProcessor()
	target.SetHashById(uint64(1), &HashResult{hash: "test", err: nil})
	res, err := target.GetHashById(uint64(2))
	if err == nil {
		t.Errorf("Error expected for missing Id")
	}
	if res != nil {
		t.Errorf("No result expected for missing Id")
	}
}

func TestTrackTime(t *testing.T) {
	target := CreateHashProcessor()
	start := time.Now()
	time.Sleep(5 * time.Millisecond)
	target.TrackTime(start)
	start = time.Now()
	time.Sleep(5 * time.Millisecond)
	target.TrackTime(start)
	start = time.Now()
	time.Sleep(5 * time.Millisecond)
	target.TrackTime(start)

	if target.stats.Total != 3 {
		t.Errorf("Expected 3 total in stats")
	}
	// A very approximate way to do this
	// Pretty sure there are better ways to accomplish that needs to be researched
	if target.stats.Average < 5000000 || target.stats.Average > 6000000 {
		t.Errorf("Time average outside of expected bounds (5-6) %d", target.stats.Average)
	}
}

func TestStartAcceptingWork(t *testing.T) {
	target := CreateHashProcessor()
	go target.StartAcceptingWork()
	// There is probably a better way to do this
	time.Sleep(100 * time.Millisecond)
	if target.jobs == nil {
		t.Errorf("worker jobs channel has not been initialized")
	}
}

func TestDoWork(t *testing.T) {
	target := CreateHashProcessor()
	go target.StartAcceptingWork()
	// There is probably a better way to do this
	time.Sleep(100 * time.Millisecond)
	target.DoWork(&HashRequest{id: uint64(1), password: "test"})
}

func TestStopAcceptingWork(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected a panic")
		}
	}()
	target := CreateHashProcessor()
	go target.StartAcceptingWork()
	// There is probably a better way to do this
	time.Sleep(100 * time.Millisecond)
	target.StopAcceptingWork()
	target.DoWork(&HashRequest{id: uint64(1), password: "test"})
}
