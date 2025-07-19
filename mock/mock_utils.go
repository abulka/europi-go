package mock

import (
	"europi/controls"
	// "europi/logutil"
	"time"
)

// Helper functions for mock input, since SetValue and SetPressed do not exist on the IKnob and IButton interfaces.
// But they do exist on the mock implementations used in tests.
func SetKnobValue(knob controls.IKnob, v int) {
	if mock, ok := knob.(interface{ SetValue(int) }); ok {
		mock.SetValue(v)
	}
}

func ButtonPress(btn controls.IButton) {
	SetButtonPressed(btn, true)
	time.Sleep(200 * time.Millisecond)
	SetButtonPressed(btn, false)
	time.Sleep(200 * time.Millisecond)
}

func SetButtonPressed(btn controls.IButton, pressed bool) {
	if mock, ok := btn.(interface{ SetPressed(bool) }); ok {
		mock.SetPressed(pressed)
	}
}

func SetDigitalInputValue(din controls.IDigitalInput, value bool) {
	if mock, ok := din.(interface{ SetValue(bool) }); ok {
		mock.SetValue(value)
	}
}

func SetAnalogueInputValue(ain controls.IAnalogueInput, volts float64) {
	if mock, ok := ain.(interface{ SetVolts(float64) }); ok {
		mock.SetVolts(volts)
	}
}

func ExitToMainMenu(hw *controls.Controls) {
	// Exit the app after 5 seconds by pressing B1 and B2 simultaneously
	SetButtonPressed(hw.B1, true)
	SetButtonPressed(hw.B2, true)
	time.Sleep(3 * time.Second) // Simulate holding both buttons for > 2 seconds to exit
	SetButtonPressed(hw.B1, false)
	SetButtonPressed(hw.B2, false)
	time.Sleep(2 * time.Second) // Wait for splash screen to clear
}

// TurnKnobToMenuIndex simulates turning the knob from the current menu index to the target index.
// thresholdPerItem is the knob delta required to move one menu item (default: 3).
func TurnKnobToMenuIndex(knob controls.IKnob, currentIndex, targetIndex, thresholdPerItem int, delay time.Duration) int {
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
func SelectMenuItem(knob controls.IKnob, idx int) {
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
	// logutil.Println("Setting knob value to", knobValue, "for menu item", idx, "of", numMenuItems)
	SetKnobValue(knob, knobValue)
}
