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
}

func NewBufferedDisplayWithFont(real IOledDevice, tinyFont bool) *BufferedDisplay {
	numLines := 3
	if tinyFont {
		numLines = 4 // TinyFont has 4 lines
	}
	return &BufferedDisplay{
		Backend:                  real,
		Lines:                    make([]string, numLines),
		highlighted:              make([]bool, numLines),
		dirty:                    false,
		lastDisplayedLines:       make([]string, numLines),
		lastDisplayedHighlighted: make([]bool, numLines),
	}
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
	m.Backend.ClearDisplay() // Clear the display first to avoid artifacts
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
