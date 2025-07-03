// IO struct and interfaces (IKnob, IButton, etc.)
package controls

import "europi/display"

// IKnob interface
// Returns the current value of the knob
// Implemented by both real and mock knobs
type IKnob interface {
	Value() int
}

type IButton interface {
	Pressed() bool
}

type IDigitalInput interface {
	Get() bool
}

type IAnalogueInput interface {
	Volts() float64
	Value() int
}

type ICV interface {
	Set(value uint32)
	On()
	Off()
}

// Controls struct holds all hardware/mocked Controls for EuroPi
// This is used by both production and mock entry points
type Controls struct {
	// Inputs
	K1, K2 IKnob
	B1, B2 IButton
	DIN    IDigitalInput
	AIN    IAnalogueInput
	// Outputs
	CV1, CV2, CV3, CV4, CV5, CV6 ICV
	Display                      display.IOledDevice
}
