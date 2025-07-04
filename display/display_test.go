// Display-related tests
package display

import (
	"testing"
)

func TestDisplayBasics(t *testing.T) {
	oled := NewMockOledDeviceWithFont(false)
	if oled == nil {
		t.Error("NewMockOledDeviceWithFont returned nil")
		return
	}
	if len(oled.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(oled.Lines))
		return
	}
	if oled.LineLen != 16 {
		t.Errorf("Expected line length 16, got %d", oled.LineLen)
		return
	}

	oled.ClearDisplay()
	
	oled.WriteLine(0, "Hello")
	if oled.Lines[0] != "Hello" {
		t.Errorf("Expected line 0 to be 'Hello', got '%s'", oled.Lines[0])
		return
	}
	oled.WriteLine(1, "World")
	if oled.Lines[1] != "World" {
		t.Errorf("Expected line 1 to be 'World', got '%s'", oled.Lines[1])
		return
	}
	oled.WriteLine(2, "Test")
	if oled.Lines[2] != "Test" {
		t.Errorf("Expected line 2 to be 'Test', got '%s'", oled.Lines[2])
		return
	}

}

func TestDisplay(t *testing.T) {
	oled := NewMockOledDeviceWithFont(false)
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
	oled.HighlightLn(1)
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
	oled.HighlightLn(-1) // Clear highlight
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
}

// TestDisplayTinyFont tests the mock OLED device with TinyFont
func TestDisplayHighlightingMutuallyExclusive(t *testing.T) {
	oled := NewMockOledDeviceWithFont(false)
	oled.WriteLine(0, "Hello")
	oled.WriteLine(1, "World")
	oled.WriteLine(2, "Test")
	oled.HighlightLn(1)
	str := oled.DisplayString()
	expected := trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World *                  │
	│Test                     │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected highlighted display output to be:\n%s\nGot:\n%s", expected, str)
	}
	oled.HighlightLn(2) // Highlight another line
	str = oled.DisplayString()
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
	oled.HighlightLn(-1) // Clear highlight
	str = oled.DisplayString()
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
}
