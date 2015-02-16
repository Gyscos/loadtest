package main

import (
	"fmt"
	"io"
	"math"
	"time"
)

type DurationSlice []time.Duration

// Forward request for length
func (s DurationSlice) Len() int {
	return len(s)
}

// Define compare
func (s DurationSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

// Define swap over an array
func (s DurationSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func showStats(times []time.Duration, w io.Writer) {
	if len(times) == 0 {
		fmt.Fprintln(w, "No times reported.")
		return
	}
	avg := 0 * time.Second
	min := 9 * time.Hour
	max := 0 * time.Second
	for _, t := range times {
		avg += t
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
	}
	avg = avg / time.Duration(len(times))

	variance := 0 * time.Second
	for _, t := range times {
		// Having squared nanoseconds may be a bit too high...
		delta := (t - avg) / time.Millisecond
		variance += delta * delta
	}
	if variance < 0 {
		panic("Negative variance???")
	}
	variance = variance / time.Duration(len(times))
	stdDev := time.Duration(math.Sqrt(float64(variance))) * time.Millisecond

	fmt.Fprintln(w, "Average time:", avg)
	fmt.Fprintln(w, "Min,max:", min, max)
	fmt.Fprintln(w, "Stddev:", stdDev)
}
