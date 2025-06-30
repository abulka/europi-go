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

func ExitToMainMenu(iox *hw.Controls) {
	// Exit the app after 5 seconds by pressing B1 and B2 simultaneously
	SetButtonPressed(iox.B1, true)
	SetButtonPressed(iox.B2, true)
	time.Sleep(3 * time.Second) // Simulate holding both buttons for > 2 seconds to exit
	SetButtonPressed(iox.B1, false)
	SetButtonPressed(iox.B2, false)
	time.Sleep(2 * time.Second) // Wait for splash screen to clear
}
