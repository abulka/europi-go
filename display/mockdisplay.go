// Mock display for host/testing
//go:build !tinygo

package display

import "bytes"

type MockOledDevice struct {
	Lines       []string
	LineLen     int // max chars per line (16 for 8x8, 21 for TinyFont)
	highlighted int // -1 for none
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
	return &MockOledDevice{LineLen: lineLen, Lines: make([]string, numLines), highlighted: -1}
}

func (m *MockOledDevice) ClearDisplay() {
	for i := range m.Lines {
		m.Lines[i] = ""
	}
}

func (m *MockOledDevice) WriteLine(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(m.Lines) {
		return // ignore out of range
	}
	// Truncate text to max line length
	if len(text) > m.LineLen {
		text = text[:m.LineLen]
	}
	m.Lines[lineNum] = text
}

func (m *MockOledDevice) DisplayString() string {
	const width = 25
	top := "┌" + string(bytes.Repeat([]byte("─"), width)) + "┐"
	bottom := "└" + string(bytes.Repeat([]byte("─"), width)) + "┘"
	var out bytes.Buffer
	out.WriteString(top + "\n")
	for i, line := range m.Lines {
		if i == m.highlighted && m.highlighted >= 0 {
			line += " *"
		}
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

func (m *MockOledDevice) HighlightLn(lineNum int) {
	m.highlighted = lineNum
}
