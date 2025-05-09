package rules

import (
	"math"
	"sync"
)

type ZScoreDetector struct {
	window    []float64
	maxSize   int
	mu        sync.Mutex
	Threshold float64
}

func NewZScoreDetector(windowSize int, threshold float64) *ZScoreDetector {
	return &ZScoreDetector{
		window:    make([]float64, 0, windowSize),
		maxSize:   windowSize,
		Threshold: threshold,
	}
}

// add a new data point and returns true if its an anomaly
func (z *ZScoreDetector) Add(value float64) bool {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.window = append(z.window, value)
	if len(z.window) > z.maxSize {
		z.window = z.window[1:]
	}

	// compute mean and std dev
	var sum, sqSum float64
	for _, v := range z.window {
		sum += v
		sqSum += v * v
	}

	n := float64(len(z.window))

	if n < 2 {
		return false // not enough data
	}
	mean := sum / n
	variance := (sqSum / n) - (mean * mean)
	stdev := math.Sqrt(variance)
	if stdev == 0 {
		return false
	}

	zScore := math.Abs(value-mean) / stdev
	return zScore > z.Threshold
}
