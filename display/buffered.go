package display

// Decorator for any IOledDevice, adding buffering capabilities
// This allows us to track changes to lines and only update the display when necessary,
// improving performance and reducing flicker.

// We prevent calls to the underlying OledHighlighter methods:
// - WriteLine if the line text is unchanged (we probe the underlying Lines slice)
// - WriteLineHighlighted if the line text is unchanged (we probe the underlying Lines slice)
// - ClearHighlight if no highlight is set
// - ClearDisplay if the display is already cleared
// - Display if no changes have been made since the last display update
// - DisplayString if no changes have been made since the last display update

type BufferedDisplay struct {
	Backend IOledDevice   // interface field: not embedded!
	Lines   []string // Lines to display, used for buffering
	highlighted []bool // Which lines have been called with WriteLineHighlighted
	dirty bool
}

func NewBufferedDisplayWithFont(real IOledDevice, tinyFont bool) *BufferedDisplay {
	numLines := 3
	if tinyFont {
		numLines = 4 // TinyFont has 4 lines
	}
	return &BufferedDisplay{
		Backend: real,
		Lines:   make([]string, numLines),
		highlighted: make([]bool, numLines),
		dirty:   false,
	}
}

func (m *BufferedDisplay) WriteLine(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(m.Lines) {
		return // ignore out of range
	}
	if m.Lines[lineNum] == text && !m.highlighted[lineNum] {
		return // No change, nothing to do
	}
	m.Backend.WriteLine(lineNum, text)
	m.Lines[lineNum] = text // Update the buffered line
	m.highlighted[lineNum] = false // Reset highlight state for this line
	m.dirty = true
}

func (m *BufferedDisplay) WriteLineHighlighted(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(m.Lines) {
		return // ignore out of range
	}
	if m.Lines[lineNum] == text && m.highlighted[lineNum] {
		return // No change, nothing to do
	}
	m.Backend.WriteLineHighlighted(lineNum, text)
	m.Lines[lineNum] = text // Update the buffered line
	m.highlighted[lineNum] = true // Set highlight state for this line
	m.dirty = true
}

func (m *BufferedDisplay) ClearDisplay() {
	// check if already in a cleared state
	allEmpty := true
	for _, line := range m.Lines {
		if line != "" {
			allEmpty = false
			break
		}
	}
	allUnhighlighted := true
	for _, highlighted := range m.highlighted {
		if highlighted {
			allUnhighlighted = false
			break
		}
	}
	// If no lines are set and no highlights, we can skip clearing
	if allEmpty && allUnhighlighted {
		return // Already cleared
	}

	m.Backend.ClearDisplay()

	// Reset lines and highlight state
	for i := range m.Lines {
		m.Lines[i] = "" // Clear all lines
		m.highlighted[i] = false // Reset highlight state
	}
	m.dirty = true
}

// For testing purposes
func (m *BufferedDisplay) DisplayString() string {
	if !m.dirty {
		return ""
	}
	// if the backend has a DisplayString method, use it
	if displayStringer, ok := m.Backend.(interface{ DisplayString() string }); ok {
		result := displayStringer.DisplayString() // Return current display state
		m.dirty = false                             // Reset dirty flag after getting display string
		return result
	}
	return "" // Fallback if no DisplayString method is available
}

func (m *BufferedDisplay) Display() {
	if !m.dirty {
		return // Nothing to update
	}
	m.Backend.Display()
	m.dirty = false // Reset dirty flag after display update
}
