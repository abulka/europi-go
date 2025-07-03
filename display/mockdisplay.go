// Mock display for host/testing
//go:build !tinygo

package display

import "bytes"

type MockOledDevice struct {
	Lines       [3]string
	Cleared     bool
	Displayed   bool
	LineLen     int // max chars per line (16 for 8x8, 21 for TinyFont)
	highlighted int // -1 for none
}

func NewMockOledDeviceWithFont(tinyFont bool) *MockOledDevice {
	lineLen := 16
	if tinyFont {
		lineLen = 21
	}
	return &MockOledDevice{LineLen: lineLen}
}

func (m *MockOledDevice) ClearDisplay() {
	m.Cleared = true
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

func (m *MockOledDevice) Display() {
	m.Displayed = true
	const width = 25
	top := "┌" + string(bytes.Repeat([]byte("─"), width)) + "┐"
	bottom := "└" + string(bytes.Repeat([]byte("─"), width)) + "┘"
	println(top)
	for i, line := range m.Lines {
		if i == m.highlighted && m.highlighted >= 0 {
			// Append highlight symbol to highlighted line
			line += " *"
		}
		println("│" + padOrTruncate(line, width) + "│")
	}
	println(bottom)
}

func padOrTruncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s + string(bytes.Repeat([]byte{' '}, n-len(s)))
}

func NewMockOledDevice() *MockOledDevice {
	return &MockOledDevice{}
}

func (m *MockOledDevice) HighlightLn(lineNum int) {
	m.highlighted = lineNum
}
