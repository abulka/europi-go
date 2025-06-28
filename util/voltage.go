package util

import "time"

// VoltageReader handles ADC to voltage conversion with calibration
type VoltageReader struct {
	MinRaw   int
	MaxRaw   int
	Scale    float64
	Filter   *AnalogFilter
	Hyst     *HysteresisFilter
	LastShow time.Time
}

func NewVoltageReader(minRaw, maxRaw int, filterSize, hystThreshold int) *VoltageReader {
	return &VoltageReader{
		MinRaw: minRaw,
		MaxRaw: maxRaw,
		Scale:  5.0 / float64(maxRaw-minRaw),
		Filter: NewAnalogFilter(filterSize),
		Hyst: &HysteresisFilter{
			Threshold: hystThreshold,
			Min:       0,
			Max:       500, // 5.00V * 100
		},
	}
}

func (v *VoltageReader) Read(raw int) float64 {
	clamped := Clamp(raw, v.MinRaw, v.MaxRaw)
	scaled := float64(clamped-v.MinRaw) * v.Scale
	filtered := v.Filter.Update(int(scaled * 100))
	stable := v.Hyst.Update(filtered)
	return float64(stable) / 100.0
}
