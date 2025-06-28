package util

import "time"

// SmartKnobProcessor with smart locking and filtering
//   - SmartKnobProcessor processes raw knob values, applies filtering, and manages locking behavior
//
// A low-pass filter (AnalogFilter) to smooth ADC readings. A time-based lock
// that filters out small jitters after inactivity. A resumeThreshold to avoid
// resuming for tiny bumps.
type SmartKnobProcessor struct {
	Filter           *AnalogFilter
	LastMapped       int
	LockedValue      int
	IsLocked         bool
	LockAfter        time.Duration // idle time to lock
	ResumeThreshold  int
	LastActivityTime time.Time
}

func NewSmartKnobProcessor() *SmartKnobProcessor {
	return &SmartKnobProcessor{
		Filter:           NewAnalogFilter(8), // To smooth more, increase window size
		LastMapped:       -1,
		LockedValue:      -1,
		IsLocked:         false,
		LockAfter:        500 * time.Millisecond,
		ResumeThreshold:  2,
		LastActivityTime: time.Now(),
	}
}

func (k *SmartKnobProcessor) Process(rawValue int) int {
	filtered := k.Filter.Update(rawValue)
	mapped := 99 - CalibrateKnobValue(filtered, 0, 65535, 0, 99)
	now := time.Now()

	if k.LastMapped == -1 {
		k.LastMapped = mapped
		k.LockedValue = mapped
		k.LastActivityTime = now
		return mapped
	}

	if mapped != k.LastMapped {
		k.LastActivityTime = now
		if k.IsLocked {
			if Abs(mapped-k.LockedValue) >= k.ResumeThreshold {
				k.IsLocked = false
				k.LockedValue = mapped
			}
		} else {
			k.LockedValue = mapped
		}
		k.LastMapped = mapped
	}

	if !k.IsLocked && now.Sub(k.LastActivityTime) > k.LockAfter {
		k.IsLocked = true
	}

	return k.LockedValue
}

// CalibrateKnobValue maps raw ADC values to a range with deadzones at extremes.
func CalibrateKnobValue(raw, inMin, inMax, outMin, outMax int) int {
	deadzone := 16
	if raw <= inMin+deadzone {
		return outMin
	}
	if raw >= inMax-deadzone {
		return outMax
	}
	scale := float64(outMax-outMin) / float64(inMax-inMin-2*deadzone)
	val := outMin + int(float64(raw-inMin-deadzone)*scale+0.5)
	return Clamp(val, outMin, outMax)
}
