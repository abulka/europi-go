//go:build !tinygo

package display

import (
	"bytes"
	"europi/logutil"

	tea "github.com/charmbracelet/bubbletea"
)

const HighlightSymbol = " *"

type MockOledDeviceTea struct {
	lines       [3]string
	program     *tea.Program
	LineLen     int // max chars per line (16 for 8x8, 21 for TinyFont)
	highlighted int // -1 for none
}

func NewMockOledDeviceTea() *MockOledDeviceTea {
	m := &MockOledDeviceTea{LineLen: 16}
	m.program = tea.NewProgram(&oledModel{lines: m.lines})
	go func() { _, _ = m.program.Run() }()
	return m
}

func NewMockOledDeviceTeaWithFont(tinyFont bool) *MockOledDeviceTea {
	lineLen := 16
	if tinyFont {
		lineLen = 21
	}
	m := &MockOledDeviceTea{LineLen: lineLen}
	m.program = tea.NewProgram(&oledModel{lines: m.lines})
	go func() { _, _ = m.program.Run() }()
	return m
}

func (m *MockOledDeviceTea) ClearDisplay() {
	m.lines = [3]string{"", "", ""}
	m.highlighted = -1
	m.update()
}

func (m *MockOledDeviceTea) WriteLine(x, y int16, text string) {
	var idx int
	switch y {
	case 10:
		idx = 0
	case 20:
		idx = 1
	case 30:
		idx = 2
	default:
		return
	}
	start := 0
	if x >= 75 {
		start = 12
	}
	line := m.lines[idx]
	if len(line) < start {
		line += makeSpacesTea(start - len(line))
	}
	// Truncate text to max line length
	if len(text) > m.LineLen {
		text = text[:m.LineLen]
	}
	if start+len(text) > len(line) {
		line = line[:start] + text
	} else {
		line = line[:start] + text + line[start+len(text):]
	}
	m.lines[idx] = line
	m.update()
}

func (m *MockOledDeviceTea) Display() {
	m.update()
	// Log OLED lines to logutil as a Unicode box with fixed width
	const width = 25
	top := "┌" + string(bytes.Repeat([]byte("─"), width)) + "┐"
	bottom := "└" + string(bytes.Repeat([]byte("─"), width)) + "┘"
	logutil.Println(top)
	for _, line := range m.lines {
		logutil.Println("│" + padOrTruncateTea(line, width) + "│")
	}
	logutil.Println(bottom)
}

func (m *MockOledDeviceTea) HighlightLn(lineNum int) {
	m.highlighted = lineNum
}

func (m *MockOledDeviceTea) update() {
	if m.program != nil {
		m.program.Send(updateMsg{lines: m.lines, highlighted: m.highlighted})
	}
}

type updateMsg struct {
	lines       [3]string
	highlighted int
}

type oledModel struct {
	lines       [3]string
	highlighted int
}

func (m *oledModel) Init() tea.Cmd {
	return nil
}

func (m *oledModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case updateMsg:
		m.lines = msg.lines
		m.highlighted = msg.highlighted
	}
	return m, nil
}

func (m *oledModel) Highlighted() int {
	return m.highlighted
}

func (m *oledModel) View() string {
	// Fixed width for border
	const width = 25
	top := "┌" + string(bytes.Repeat([]byte("─"), width)) + "┐"
	bottom := "└" + string(bytes.Repeat([]byte("─"), width)) + "┘"
	var b bytes.Buffer
	b.WriteString(top + "\n")
	for i, line := range m.lines {
		toShow := line
		if m.highlighted == i && m.highlighted >= 0 && len(line) > 0 {
			toShow = line + HighlightSymbol
		}
		b.WriteString("│" + padOrTruncateTea(toShow, width) + "│\n")
	}
	b.WriteString(bottom)
	return b.String()
}

func makeSpacesTea(n int) string {
	if n <= 0 {
		return ""
	}
	return string(bytes.Repeat([]byte{' '}, n))
}

func padOrTruncateTea(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s + string(bytes.Repeat([]byte{' '}, n-len(s)))
}
