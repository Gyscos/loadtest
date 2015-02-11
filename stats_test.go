package main

import (
	"os"
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	times := []time.Duration{
		5 * time.Second,
		4 * time.Minute,
		5 * time.Second,
		6 * time.Second,
		5 * time.Second,
		4 * time.Second,
		5 * time.Second,
		4 * time.Second,
		6 * time.Second,
		5 * time.Second,
		4 * time.Second,
		5 * time.Second,
		6 * time.Second,
		5 * time.Second,
		4 * time.Second,
		5 * time.Second,
		4 * time.Second,
		6 * time.Second,
	}
	showStats(times, os.Stdout)
}
