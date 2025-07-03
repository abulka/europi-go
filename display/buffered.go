package display

// BufferedOledDevice wraps an IOledDevice and provides line buffering, change detection, and highlight state.
type BufferedOledDevice struct {
	inner         IOledDevice
	lines         []string
	prevLines     []string
	highlighted   int
	prevHighlight int
	lineHeight    int16
	startY        int16
}

// NewBufferedOledDevice creates a new buffered OLED device
func NewBufferedOledDevice(inner IOledDevice, lineCount int, startY, lineHeight int16) *BufferedOledDevice {
	return &BufferedOledDevice{
		inner:         inner,
		lines:         make([]string, lineCount),
		prevLines:     make([]string, lineCount),
		highlighted:   -1,
		prevHighlight: -1,
		lineHeight:    lineHeight,
		startY:        startY,
	}
}

func (b *BufferedOledDevice) ClearDisplay() {
	for i := range b.lines {
		b.lines[i] = ""
	}
	b.highlighted = -1
	b.prevHighlight = -1
	b.inner.ClearDisplay()
	b.prevLines = make([]string, len(b.lines))
}

// Writeln writes a line to the buffer (doesn't display immediately)
func (b *BufferedOledDevice) Writeln(lineNum int, text string) {
	if lineNum >= 0 && lineNum < len(b.lines) {
		if b.lines[lineNum] != text {
			b.lines[lineNum] = text
		}
	}
}

func (b *BufferedOledDevice) WriteLine(x, y int16, text string) {
	lineNum := int((y - b.startY) / b.lineHeight)
	b.Writeln(lineNum, text)
}

func (b *BufferedOledDevice) HighlightLn(lineNum int) {
	b.highlighted = lineNum
	b.inner.HighlightLn(lineNum)
}

func (b *BufferedOledDevice) hasChanges() bool {
	for i := range b.lines {
		if b.lines[i] != b.prevLines[i] {
			return true
		}
	}
	if b.highlighted != b.prevHighlight {
		return true
	}
	return false
}

func (b *BufferedOledDevice) Display() {
	if !b.hasChanges() {
		return
	}
	b.inner.ClearDisplay()
	// Set highlight ONCE for the whole display
	b.inner.HighlightLn(b.highlighted)
	for i := 0; i < len(b.lines); i++ {
		y := b.startY + int16(i)*b.lineHeight
		b.inner.WriteLine(0, y, b.lines[i])
	}
	b.inner.Display()
	copy(b.prevLines, b.lines)
	b.prevHighlight = b.highlighted
}
