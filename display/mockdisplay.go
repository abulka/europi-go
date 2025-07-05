// Mock display for host/testing
//go:build !tinygo

package display

import "bytes"

type MockOledDevice struct {
	LinesRaw []string // like a real OLED, but in memory
	LineLen  int      // max chars per line (16 for 8x8, 21 for TinyFont)
}

func NewMockOledDeviceWithFont(tinyFont bool) *MockOledDevice {
	lineLen := 16
	if tinyFont {
		lineLen = 21
	}
	numLines := 3
	if tinyFont {
		numLines = 4 // TinyFont has 4 lines
	}
	return &MockOledDevice{LineLen: lineLen, LinesRaw: make([]string, numLines)}
}

func (m *MockOledDevice) ClearDisplay() {
	for i := range m.LinesRaw {
		m.LinesRaw[i] = ""
	}
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
