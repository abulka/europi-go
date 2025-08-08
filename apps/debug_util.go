package apps

// findCVIndex returns the index of the given pulse in the pulses slice.
func findCVIndex(pulses []*PulseOutput, target *PulseOutput) int {
	for i, p := range pulses {
		if p == target {
			return i
		}
	}
	return -1
}
