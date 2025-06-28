// Mock display for host/testing
//go:build !tinygo

package display

import "bytes"

type MockOledDevice struct {
	Lines     [3]string // 3 lines for y=10,20,30
	Cleared   bool
	Displayed bool
}

func (m *MockOledDevice) ClearDisplay() {
	m.Cleared = true
	for i := range m.Lines {
		m.Lines[i] = ""
	}
}

func (m *MockOledDevice) WriteLine(x, y int16, text string) {
	// Map y to line index: y=10 -> 0, y=20 -> 1, y=30 -> 2
	var idx int
	switch y {
	case 10:
		idx = 0
	case 20:
		idx = 1
	case 30:
		idx = 2
	default:
		return // ignore lines not mapped
	}
	// x=0 is start of line, x=75 is about 12 chars in
	start := 0
	if x >= 75 {
		start = 12
	}
	// Pad line if needed
	line := m.Lines[idx]
	if len(line) < start {
		line += makeSpaces(start - len(line))
	}
	// Insert/replace text at position
	if start+len(text) > len(line) {
		line = line[:start] + text
	} else {
		line = line[:start] + text + line[start+len(text):]
	}
	m.Lines[idx] = line
}

func (m *MockOledDevice) Display() {
	m.Displayed = true
	const width = 25
	top := "┌" + string(bytes.Repeat([]byte("─"), width)) + "┐"
	bottom := "└" + string(bytes.Repeat([]byte("─"), width)) + "┘"
	println(top)
	for _, line := range m.Lines {
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

func makeSpaces(n int) string {
	if n <= 0 {
		return ""
	}
	return string(bytes.Repeat([]byte{' '}, n))
}
