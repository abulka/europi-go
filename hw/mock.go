//go:build !tinygo

// Mock implementations (pure Go, no hardware deps)
package hw

import "europi/display"

// MockKnob implements IKnob
// SetValue allows test code to set the value
// Value returns the current value

type MockKnob struct{ val int }

func (m *MockKnob) Value() int     { return m.val }
func (m *MockKnob) SetValue(v int) { m.val = v }

// MockButton implements IButton
// SetPressed allows test code to set the pressed state

type MockButton struct{ pressed bool }

func (m *MockButton) Pressed() bool     { return m.pressed }
func (m *MockButton) SetPressed(p bool) { m.pressed = p }

// MockDigitalInput implements IDigitalInput
// SetState allows test code to set the state

type MockDigitalInput struct{ state bool }

func (m *MockDigitalInput) Get() bool       { return m.state }
func (m *MockDigitalInput) SetState(s bool) { m.state = s }

// MockAnalogueInput implements IAnalogueInput
// SetVolts/SetValue allow test code to set the values

type MockAnalogueInput struct {
	volts float64
	value int
}

func (m *MockAnalogueInput) Volts() float64     { return m.volts }
func (m *MockAnalogueInput) Value() int         { return m.value }
func (m *MockAnalogueInput) SetVolts(v float64) { m.volts = v }
func (m *MockAnalogueInput) SetValue(val int)   { m.value = val }

// MockCV implements ICV

type MockCV struct {
	val     uint32
	on, off bool
}

func (m *MockCV) Set(v uint32) { m.val = v }
func (m *MockCV) On()          { m.on = true }
func (m *MockCV) Off()         { m.off = true }

// SetupEuroPiMock returns an IO struct with all fields set to mocks
func SetupEuroPiMock() *IO {
	return &IO{
		K1:      &MockKnob{},
		K2:      &MockKnob{},
		B1:      &MockButton{},
		B2:      &MockButton{},
		DIN:     &MockDigitalInput{},
		AIN:     &MockAnalogueInput{},
		CV1:     &MockCV{},
		CV2:     &MockCV{},
		CV3:     &MockCV{},
		CV4:     &MockCV{},
		CV5:     &MockCV{},
		CV6:     &MockCV{},
		Display: &display.MockOledDevice{}, // Added mock display
	}
}
