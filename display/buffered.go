package display

// BufferedOledDevice wraps an IOledDevice and provides line buffering with dirty line tracking.
type BufferedOledDevice struct {
	inner       IOledDevice
	lines       []string
	dirty       []bool
	highlighted int
	lineCount   int
}

// NewBufferedOledDevice creates a new buffered OLED device
func NewBufferedOledDevice(inner IOledDevice, lineCount int) *BufferedOledDevice {
	return &BufferedOledDevice{
		inner:       inner,
		lines:       make([]string, lineCount),
		dirty:       make([]bool, lineCount),
		highlighted: -1,
		lineCount:   lineCount,
	}
}

func (b *BufferedOledDevice) ClearDisplay() {
	// Clear all lines and mark them as dirty
	for i := range b.lines {
		b.lines[i] = ""
		b.dirty[i] = true
	}
	b.highlighted = -1
	b.inner.ClearDisplay()
}

func (b *BufferedOledDevice) WriteLine(lineNum int, text string) {
	if lineNum < 0 || lineNum >= b.lineCount {
		return // Invalid line number
	}
	// Only mark as dirty if the content actually changed
	if b.lines[lineNum] != text {
		b.lines[lineNum] = text
		b.dirty[lineNum] = true
	}
}

func (b *BufferedOledDevice) HighlightLn(lineNum int) {
	if b.highlighted == lineNum {
		return // No change needed
	}

	// Mark old highlighted line as dirty (if valid)
	if b.highlighted >= 0 && b.highlighted < b.lineCount {
		b.dirty[b.highlighted] = true
	}

	// Mark new highlighted line as dirty (if valid)
	if lineNum >= 0 && lineNum < b.lineCount {
		b.dirty[lineNum] = true
	}

	b.highlighted = lineNum
	b.inner.HighlightLn(lineNum)
}


func (b *BufferedOledDevice) Display() {
	// Check if any lines are dirty
	hasDirtyLines := false
	for i := range b.dirty {
		if b.dirty[i] {
			hasDirtyLines = true
			break
		}
	}

	if !hasDirtyLines {
		return // Nothing to update
	}

	// Clear display and redraw ALL lines (not just dirty ones)
	b.inner.ClearDisplay()
	b.inner.HighlightLn(b.highlighted)

	// Write all lines to the inner device
	for i := 0; i < b.lineCount; i++ {
		b.inner.WriteLine(i, b.lines[i])
	}

	// Clear all dirty flags
	for i := range b.dirty {
		b.dirty[i] = false
	}

	b.inner.Display()
}