// IO struct and interfaces (IKnob, IButton, etc.)
package hw

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

type IOledDevice interface {
	ClearDisplay()
	Display()
	WriteLine(x, y int16, text string)
}

// IO struct holds all hardware/mocked IO for EuroPi
// This is used by both production and mock entry points
type IO struct {
	// Inputs
	K1, K2 IKnob
	B1, B2 IButton
	DIN    IDigitalInput
	AIN    IAnalogueInput
	// Outputs
	CV1, CV2, CV3, CV4, CV5, CV6 ICV
	Display                      IOledDevice
}
