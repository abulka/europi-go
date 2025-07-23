// Unit tests for IO logic/mocks
package controls

import "testing"

func TestIO(t *testing.T) {
	// TODO: Add IO tests
}

// TestControls should be a top-level function, not nested
func TestKnobChoice(t *testing.T) {
	// Create a mock Controls instance
	mockControls := &Controls{
		K1: &MockKnob{},
		K2: &MockKnob{},
	}

	// Test the Choice method
	k1, ok := mockControls.K1.(*MockKnob)
	if ok {
		// TODO aren't we supposed to be 0..100 range?
		k1.SetValue(32767) // Set a value for testing
	} else {
		t.Errorf("K1 is not a MockKnob")
	}
	choice := mockControls.K1.Choice([]int{0, 50, 100})
	if choice != 50 {
		t.Errorf("Expected 50, got %d", choice)
	}
}
