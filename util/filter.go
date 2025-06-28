// Package util provides pure logic utilities for EuroPi
package util

// abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// clamp restricts x to the range [min, max].
func Clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// AnalogFilter with configurable size and range calibration - A "low-pass filter" perhaps
//   - A filter averages out rapid fluctuations that might otherwise trigger hysteresis
type AnalogFilter struct {
	history  []int // Slice for flexible window size
	index    int
	sum      int
	capacity int
}

func NewAnalogFilter(windowSize int) *AnalogFilter {
	return &AnalogFilter{
		history:  make([]int, windowSize),
		capacity: windowSize,
	}
}

func (f *AnalogFilter) Update(value int) int {
	f.sum -= f.history[f.index]
	f.history[f.index] = value
	f.sum += value
	f.index = (f.index + 1) % len(f.history)
	return f.sum / len(f.history)
}

// HysteresisFilter with edge-case handling
//   - HysteresisFilter provides final "gatekeeping" after filtering, applies a threshold to prevent rapid toggling
type HysteresisFilter struct {
	LastValue int
	Threshold int
	Min       int // Minimum expected value
	Max       int // Maximum expected value
}

func (f *HysteresisFilter) Update(value int) int {
	if value <= f.Min || value >= f.Max {
		f.LastValue = value
		return value
	}
	if Abs(value-f.LastValue) >= f.Threshold {
		f.LastValue = value
	}
	return f.LastValue
}
