// Display-related tests
package display

import (
	"europi/util"
	"testing"
)

func TestDisplayBasics(t *testing.T) {
	oled := NewMockOledDevice(3, 16)
	if oled == nil {
		t.Error("NewMockOledDevice returned nil")
		return
	}
	if len(oled.LinesRaw) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(oled.LinesRaw))
		return
	}
	if oled.LineLen != 16 {
		t.Errorf("Expected line length 16, got %d", oled.LineLen)
		return
	}

	oled.ClearDisplay()

	oled.WriteLine(0, "Hello")
	if oled.LinesRaw[0] != "Hello" {
		t.Errorf("Expected line 0 to be 'Hello', got '%s'", oled.LinesRaw[0])
		return
	}
	oled.WriteLine(1, "World")
	if oled.LinesRaw[1] != "World" {
		t.Errorf("Expected line 1 to be 'World', got '%s'", oled.LinesRaw[1])
		return
	}
	oled.WriteLine(2, "Test")
	if oled.LinesRaw[2] != "Test" {
		t.Errorf("Expected line 2 to be 'Test', got '%s'", oled.LinesRaw[2])
		return
	}

}

func TestDisplay(t *testing.T) {
	oled := NewMockOledDevice(3, 16)
	oled.ClearDisplay()
	oled.WriteLine(0, "Hello")
	oled.WriteLine(1, "World")
	oled.WriteLine(2, "Test")
	str := oled.DisplayString()
	// Check if display output is as expected
	expected := util.Trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World                    │
	│Test                     │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected display output to be:\n%s\nGot:\n%s", expected, str)
	}
	oled.WriteLineHighlighted(1, "World")
	str = oled.DisplayString()
	// Check if highlighted line is marked
	expected = util.Trimdedent(`
	┌─────────────────────────┐
	│Hello                    │
	│World *                  │
	│Test                     │
	└─────────────────────────┘
	`)
	if str != expected {
		t.Errorf("Expected highlighted display output to be:\n%s\nGot:\n%s", expected, str)
	}
	oled.WriteLine(1, "World") // Remove highlight
	str = oled.DisplayString()
	// Check if highlight is cleared
	expected = util.Trimdedent(`
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
