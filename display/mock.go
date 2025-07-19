// Mock display for host/testing
//go:build !tinygo

package display

import "bytes"

type MockOledDevice struct {
	LinesRaw []string // like a real OLED, but in memory
	LineLen  int      // max chars per line (16 for 8x8, 21 for TinyFont)
	numLines int      // number of lines (3 or 4)
}

// GetSSD1306 returns nil for the mock device.
// Its really a *ssd1306.Device but we return nil here since it's a mock
func (m *MockOledDevice) GetSSD1306() any {
	return nil
}

func NewMockOledDevice(numLines, lineLen int) *MockOledDevice {
	m := &MockOledDevice{LineLen: lineLen}
	m.SetNumLines(numLines)
	return m
}

func (m *MockOledDevice) SetNumLines(numLines int) {
	if numLines < 3 || numLines > 4 {
		panic("numLines must be 3 or 4")
	}
	m.numLines = numLines
	m.LinesRaw = make([]string, numLines) // reset lines to empty
}

func (m *MockOledDevice) NumLines() int {
	return m.numLines
}

func (m *MockOledDevice) ClearDisplay() {
	for i := range m.LinesRaw {
		m.LinesRaw[i] = ""
	}
}

func (m *MockOledDevice) ClearBuffer() {
	// In a mock, ClearBuffer is the same as ClearDisplay
	m.ClearDisplay()
}

func (m *MockOledDevice) WriteLine(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(m.LinesRaw) {
		return // ignore out of range
	}
	// Truncate text to max line length
	if len(text) > m.LineLen {
		text = text[:m.LineLen]
	}
	m.LinesRaw[lineNum] = text
}

func (m *MockOledDevice) WriteLineHighlighted(lineNum int, text string) {
	marker := " *"
	maxTextLen := m.LineLen - len(marker)
	if maxTextLen < 0 {
		maxTextLen = 0
	}
	if len(text) > maxTextLen {
		text = text[:maxTextLen]
	}
	m.LinesRaw[lineNum] = text + marker
}

func (m *MockOledDevice) DisplayString() string {
	const width = 25
	top := "┌" + string(bytes.Repeat([]byte("─"), width)) + "┐"
	bottom := "└" + string(bytes.Repeat([]byte("─"), width)) + "┘"
	var out bytes.Buffer
	out.WriteString(top + "\n")
	for _, line := range m.LinesRaw {
		out.WriteString("│" + padOrTruncate(line, width) + "│\n")
	}
	out.WriteString(bottom + "\n")
	return out.String()
}

func (m *MockOledDevice) Display() {
	print(m.DisplayString())
}

func padOrTruncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s + string(bytes.Repeat([]byte{' '}, n-len(s)))
}
