package display

// BufferedDisplay wraps an IOledDevice and only updates the underlying device when the buffer
// (intended display state) differs from the last displayed state, minimizing unnecessary redraws and flicker.
type BufferedDisplay struct {
	Backend                  IOledDevice // The actual device being decorated - interface field: not embedded!
	Lines                    []string    // Current intended display lines
	highlighted              []bool      // Current intended highlight state per line
	dirty                    bool        // True if buffer differs from last displayed state
	lastDisplayedLines       []string    // Last lines actually sent to the display
	lastDisplayedHighlighted []bool      // Last highlight state actually sent to the display
	numLines                 int         // Number of lines (3 or 4)
}

func NewBufferedDisplay(real IOledDevice, numLines int) *BufferedDisplay {
	m := &BufferedDisplay{
		Backend: real,
	}
	m.SetNumLines(numLines)
	return m
}

func (m *BufferedDisplay) GetSSD1306() any {
	return m.Backend.GetSSD1306()
}

func (m *BufferedDisplay) SetNumLines(numLines int) {
	if numLines < 3 || numLines > 4 {
		panic("numLines must be 3 or 4")
	}
	m.numLines = numLines
	m.Lines = make([]string, numLines)
	m.highlighted = make([]bool, numLines)
	m.lastDisplayedLines = make([]string, numLines)
	m.lastDisplayedHighlighted = make([]bool, numLines)

	m.Backend.SetNumLines(numLines) // Call the underlying device's SetNumLines to ensure it knows the line count
	
	m.dirty = true // Mark dirty so next display updates backend
}

// NumLines returns the number of lines (3 or 4) for the display.
func (m *BufferedDisplay) NumLines() int {
	return m.numLines
}

// WriteLine updates the buffer for the given line and removes highlight for that line.
// No backend calls are made until Display or DisplayString. Sets dirty only if the buffer differs from the last displayed state.
func (m *BufferedDisplay) WriteLine(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(m.Lines) {
		return
	}
	m.Lines[lineNum] = text
	m.highlighted[lineNum] = false
	m.dirty = !m.isBufferEqualToLastDisplayed()
}

// WriteLineHighlighted updates the buffer for the given line and sets highlight for that line.
// No backend calls are made until Display or DisplayString. Sets dirty only if the buffer differs from the last displayed state.
func (m *BufferedDisplay) WriteLineHighlighted(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(m.Lines) {
		return
	}
	m.Lines[lineNum] = text
	m.highlighted[lineNum] = true
	m.dirty = !m.isBufferEqualToLastDisplayed()
}

// ClearDisplay resets the buffer to an empty state. No backend calls are made until Display or DisplayString.
// Sets dirty only if the buffer differs from the last displayed state.
func (m *BufferedDisplay) ClearDisplay() {
	for i := range m.Lines {
		m.Lines[i] = ""
		m.highlighted[i] = false
	}
	m.dirty = !m.isBufferEqualToLastDisplayed()
}

// ClearBuffer resets the buffer to an empty state. No backend calls are made until Display or DisplayString.
// Sets dirty only if the buffer differs from the last displayed state.
func (m *BufferedDisplay) ClearBuffer() {
	// This is the same as ClearDisplay in BufferedDisplay
	m.ClearDisplay()
}

// DisplayString pushes all buffered changes to the backend mock and returns the current display as a string.
// After returning, it updates the last displayed state. Used for testing.
func (m *BufferedDisplay) DisplayString() string {
	if !m.dirty {
		return ""
	}
	m.flushBufferToBackend()
	if displayStringer, ok := m.Backend.(interface{ DisplayString() string }); ok {
		result := displayStringer.DisplayString()
		copy(m.lastDisplayedLines, m.Lines)
		copy(m.lastDisplayedHighlighted, m.highlighted)
		m.dirty = false
		return result
	}
	return ""
}

// Display pushes all buffered changes to the backend and updates the last displayed state.
func (m *BufferedDisplay) Display() {
	if !m.dirty {
		return
	}
	m.flushBufferToBackend()
	m.Backend.Display()
	copy(m.lastDisplayedLines, m.Lines)
	copy(m.lastDisplayedHighlighted, m.highlighted)
	m.dirty = false
}

// flushBufferToBackend pushes all buffered changes to the backend
func (m *BufferedDisplay) flushBufferToBackend() {
	m.Backend.ClearBuffer() // Don't use ClearDisplay() as it causes flicker
	for i := range m.Lines {
		if m.highlighted[i] {
			m.Backend.WriteLineHighlighted(i, m.Lines[i])
		} else {
			m.Backend.WriteLine(i, m.Lines[i])
		}
	}
}

// isBufferEqualToLastDisplayed returns true if the buffer matches the last displayed state.
func (m *BufferedDisplay) isBufferEqualToLastDisplayed() bool {
	if len(m.Lines) != len(m.lastDisplayedLines) || len(m.highlighted) != len(m.lastDisplayedHighlighted) {
		return false
	}
	for i := range m.Lines {
		if m.Lines[i] != m.lastDisplayedLines[i] || m.highlighted[i] != m.lastDisplayedHighlighted[i] {
			return false
		}
	}
	return true
}
