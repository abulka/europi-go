// Test buffering to see it optimises OLED display updates
package display

import (
	"testing"
)

func TestBuffered(t *testing.T) {
	// var real IOledDisplay = &SSD1306{}
	// buffered := &BufferedDisplay{Backend: real}

	var real IOledDevice = NewMockOledDeviceWithFont(false)
	var oled *BufferedDisplay = NewBufferedDisplayWithFont(real, false)
	oled.ClearDisplay()
	oled.WriteLine(0, "Hello")
	oled.WriteLine(1, "World")
	oled.WriteLine(2, "Test")
	str := oled.DisplayString()
	// Check if display output is as expected
	expected := trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World                    │
	│Test                     │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected display output to be:\n%s\nGot:\n%s", expected, str)
	}

	oled.WriteLine(0, "Hello") // No change, should not mark as dirty
	str = oled.DisplayString()
	expected = ""
	if str != expected {
		t.Errorf("Expected no change after writing same line, got:\n%s", str)
	}

	// Call display again, should not update
	str = oled.DisplayString()
	if str != expected {
		t.Errorf("Expected no change after second display call, got:\n%s", str)
	}

	oled.WriteLineHighlighted(1, "World")
	str = oled.DisplayString()
	// Check if highlighted line is marked
	expected = trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World *                  │
	│Test                     │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected highlighted display output to be:\n%s\nGot:\n%s", expected, str)
	}

	// highlight the same line again, should not change
	oled.WriteLineHighlighted(1, "World")
	str = oled.DisplayString()
	expected = ""
	if str != expected {
		t.Errorf("Expected no change after highlighting same line, got:\n%s", str)
	}

	// highlight a different line
	oled.WriteLine(1, "World") // Remove highlight
	oled.WriteLineHighlighted(2, "Test") // Highlight the third line
	str = oled.DisplayString()
	// Check if second line is now highlighted
	expected = trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World                    │
	│Test *                   │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected second highlight display output to be:\n%s\nGot:\n%s", expected, str)
	}

	// Clear highlight
	oled.WriteLine(2, "Test") // Remove highlight
	str = oled.DisplayString()
	// Check if highlight is cleared
	expected = trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World                    │
	│Test                     │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected cleared highlight display output to be:\n%s\nGot:\n%s", expected, str)
	}

	// Clear highlight again, should not change
	oled.WriteLine(2, "Test")
	str = oled.DisplayString()
	expected = ""
	if str != expected {
		t.Errorf("Expected no change after clearing highlight again, got:\n%s", str)
	}

	// Clear display
	oled.ClearDisplay()
	str = oled.DisplayString()
	expected = trimdedent(`
	┌─────────────────────────┐
	│                         │
	│                         │
	│                         │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected cleared display output to be:\n%s\nGot:\n%s", expected, str)
	}
	// Clear display again, should not change
	oled.ClearDisplay()
	str = oled.DisplayString()
	expected = ""
	if str != expected {
		t.Errorf("Expected no change after clearing display again, got:\n%s", str)
	}

	// Test using WriteLineHighlighted - bypasses the mutual exclusion logic of HighlightLn
	oled.WriteLineHighlighted(0, "Highlighted")
	str = oled.DisplayString()
	expected = trimdedent(`
	┌─────────────────────────┐
	│Highlighted *            │
	│                         │
	│                         │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected highlighted line output to be:\n%s\nGot:\n%s", expected, str)
	}
	oled.WriteLineHighlighted(1, "Highlighted2")
	str = oled.DisplayString()
	expected = trimdedent(`
	┌─────────────────────────┐
	│Highlighted *            │
	│Highlighted2 *           │
	│                         │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected multiple highlighted lines output to be:\n%s\nGot:\n%s", expected, str)
	}

	// Call DisplayString again, should not change
	str = oled.DisplayString()
	expected = ""
	if str != expected {
		t.Errorf("Expected no change after calling DisplayString again, got:\n%s", str)
	}

	// Call WriteLineHighlighted(1, "Highlighted2") again, should not change
	oled.WriteLineHighlighted(1, "Highlighted2")
	str = oled.DisplayString()
	expected = ""
	if str != expected {
		t.Errorf("Expected no change after writing same highlighted line again, got:\n%s", str)
	}
}

// Test if ClearDisplay() then we write the same content then buffered decorator
// should not trigger a full redraw when we call DisplayString() again.
func TestBufferedClearDisplay(t *testing.T) {
	real := NewMockOledDeviceWithFont(false)
	oled := NewBufferedDisplayWithFont(real, false)
	oled.ClearDisplay()
	oled.WriteLine(0, "Hello")
	oled.WriteLine(1, "World")
	oled.WriteLine(2, "Test")
	str := oled.DisplayString()
	expected := trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World                    │
	│Test                     │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected display output to be:\n%s\nGot:\n%s", expected, str)
	}

	oled.ClearDisplay() // Clear display
	// Now we write the same content again, should not trigger a full redraw
	// just because we cleared the display
	oled.WriteLine(0, "Hello")
	oled.WriteLine(1, "World")
	oled.WriteLine(2, "Test")
	str = oled.DisplayString() // Should not trigger full redraw
	expected = ""
	if str != expected {
		t.Errorf("Expected no change after clearing display, got:\n%s", str)
	}

	// Call DisplayString again, should not change
	str = oled.DisplayString()
	expected = ""
	if str != expected {
		t.Errorf("Expected no change after calling DisplayString again, got:\n%s", str)
	}

	// Now write something different
	oled.WriteLine(0, "New Line")
	str = oled.DisplayString()
	expected = trimdedent(`
	┌─────────────────────────┐
	│New Line                 │
	│World                    │
	│Test                     │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected new line display output to be:\n%s\nGot:\n%s", expected, str)
	}
}