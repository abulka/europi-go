package mock

import (
	hw "europi/controls"
	"time"
)

// Helper functions for mock input, since SetValue and SetPressed do not exist on the IKnob and IButton interfaces.
// But they do exist on the mock implementations used in tests.
func SetKnobValue(knob hw.IKnob, v int) {
	if mock, ok := knob.(interface{ SetValue(int) }); ok {
		mock.SetValue(v)
	}
}

func SetButtonPressed(btn hw.IButton, pressed bool) {
	if mock, ok := btn.(interface{ SetPressed(bool) }); ok {
		mock.SetPressed(pressed)
	}
}

func SetDigitalInputValue(din hw.IDigitalInput, value bool) {
	if mock, ok := din.(interface{ SetValue(bool) }); ok {
		mock.SetValue(value)
	}
}

func SetAnalogueInputValue(ain hw.IAnalogueInput, volts float64) {
	if mock, ok := ain.(interface{ SetVolts(float64) }); ok {
		mock.SetVolts(volts)
	}
}

func ExitToMainMenu(iox *hw.Controls) {
	// Exit the app after 5 seconds by pressing B1 and B2 simultaneously
	SetButtonPressed(iox.B1, true)
	SetButtonPressed(iox.B2, true)
	time.Sleep(3 * time.Second) // Simulate holding both buttons for > 2 seconds to exit
	SetButtonPressed(iox.B1, false)
	SetButtonPressed(iox.B2, false)
	time.Sleep(2 * time.Second) // Wait for splash screen to clear
}

// TurnKnobToMenuIndex simulates turning the knob from the current menu index to the target index.
// thresholdPerItem is the knob delta required to move one menu item (default: 3).
func TurnKnobToMenuIndex(knob hw.IKnob, currentIndex, targetIndex, thresholdPerItem int, delay time.Duration) int {
	steps := (targetIndex - currentIndex) * thresholdPerItem
	if steps == 0 {
		return currentIndex
	}
	prevValue := 0
	if mock, ok := knob.(interface{ GetValue() int }); ok {
		prevValue = mock.GetValue()
	}
	stepDir := 1
	if steps < 0 {
		stepDir = -1
	}
	for i := 0; i < abs(steps); i++ {
		newValue := prevValue + stepDir
		SetKnobValue(knob, newValue)
		prevValue = newValue
		time.Sleep(delay)
	}
	return targetIndex
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

var numMenuItems = 3

// SetNumMenuItems stores the number of menu items for SelectMenuItem to use.
func SetNumMenuItems(n int) { numMenuItems = n }

// SelectMenuItem sets the knob value so the menu highlight matches the desired index.
func SelectMenuItem(knob hw.IKnob, idx int) {
	if numMenuItems < 2 {
		SetKnobValue(knob, 0)
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= numMenuItems {
		idx = numMenuItems - 1
	}
	knobValue := int(float64(idx) * 100.0 / float64(numMenuItems-1))
	SetKnobValue(knob, knobValue)
}
