//go:build !tinygo

package display

import (
	"bytes"
	"europi/logutil"

	tea "github.com/charmbracelet/bubbletea"
)

const HighlightSymbol = " *"

type MockOledDeviceTea struct {
	LinesRaw []string // like a real OLED, but in memory
	program  *tea.Program
	LineLen  int // max chars per line (16 for 8x8, 21 for TinyFont)
}

func NewMockOledDeviceTeaWithFont(tinyFont bool) *MockOledDeviceTea {
	lineLen := 16
	if tinyFont {
		lineLen = 21
	}
	numLines := 3
	if tinyFont {
		numLines = 4
	}
	m := &MockOledDeviceTea{LineLen: lineLen, LinesRaw: make([]string, numLines)}
	m.program = tea.NewProgram(&oledModel{lines: m.LinesRaw})
	go func() { _, _ = m.program.Run() }()
	return m
}

func (m *MockOledDeviceTea) ClearDisplay() {
	for i := range m.LinesRaw {
		m.LinesRaw[i] = ""
	}
	m.update()
}

func (m *MockOledDeviceTea) WriteLine(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(m.LinesRaw) {
		return // ignore out of range
	}
	// Truncate text to max line length
	if len(text) > m.LineLen {
		text = text[:m.LineLen]
	}
	m.LinesRaw[lineNum] = text
	m.update()
}

func (m *MockOledDeviceTea) WriteLineHighlighted(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(m.LinesRaw) {
		return // ignore out of range
	}
	marker := HighlightSymbol
	maxTextLen := m.LineLen - len(marker)
	if maxTextLen < 0 {
		maxTextLen = 0
	}
	if len(text) > maxTextLen {
		text = text[:maxTextLen]
	}
	m.LinesRaw[lineNum] = text + marker
	m.update()
}

func (m *MockOledDeviceTea) Display() {
	m.update()
	logutil.Println(m.DisplayString())
}

func (m *MockOledDeviceTea) DisplayString() string {
	const width = 25
	top := "┌" + string(bytes.Repeat([]byte("─"), width)) + "┐"
	bottom := "└" + string(bytes.Repeat([]byte("─"), width)) + "┘"
	var out bytes.Buffer
	out.WriteString(top + "\n")
	for _, line := range m.LinesRaw {
		out.WriteString("│" + padOrTruncateTea(line, width) + "│\n")
	}
	out.WriteString(bottom + "\n")
	return out.String()
}

func (m *MockOledDeviceTea) update() {
	if m.program != nil {
		// Send a copy of the lines slice to avoid race conditions
		linesCopy := make([]string, len(m.LinesRaw))
		copy(linesCopy, m.LinesRaw)
		m.program.Send(updateMsg{lines: linesCopy})
	}
}

type updateMsg struct {
	lines []string
}

type oledModel struct {
	lines []string
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
	}
	return m, nil
}

func (m *oledModel) View() string {
	// Fixed width for border
	const width = 25
	top := "┌" + string(bytes.Repeat([]byte("─"), width)) + "┐"
	bottom := "└" + string(bytes.Repeat([]byte("─"), width)) + "┘"
	var b bytes.Buffer
	b.WriteString(top + "\n")
	for _, line := range m.lines {
		b.WriteString("│" + padOrTruncateTea(line, width) + "│\n")
	}
	b.WriteString(bottom)
	return b.String()
}

func padOrTruncateTea(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s + string(bytes.Repeat([]byte{' '}, n-len(s)))
}
