package util

import (
	"sync"
	"time"
)

// TimeVisualiser holds multiple lines of text with a virtual cursor and time-based advancement.
type TimeVisualiser struct {
	mu             sync.Mutex
	lines          [][]rune // simulates multi-line text display timeline
	displayBuf     []rune // convert lines to a single display buffer when outputting to channel
	lastDisplayBuf []rune  // optimization
	numLines       int
	cursor         int
	maxChars       int
	interval       time.Duration
	lastAdvance    time.Time
	displayUpdates chan string // channel for sending display updates, create a goroutine to read from this
}

// NewTimeVisualiser creates a new TimeVisualiser with fixed-size rune buffers.
func NewTimeVisualiser(numLines, maxChars int, interval time.Duration) *TimeVisualiser {
	if numLines < 1 {
		numLines = 1
	}
	if maxChars < 1 {
		maxChars = 80
	}
	lines := make([][]rune, numLines)
	for i := range lines {
		lines[i] = make([]rune, maxChars)
		for j := range lines[i] {
			lines[i][j] = ' '
		}
	}
	bufSize := numLines*(maxChars+1)
	return &TimeVisualiser{
		lines:          lines,
		numLines:       numLines,
		maxChars:       maxChars,
		interval:       interval,
		lastAdvance:    time.Now(),
		displayUpdates: make(chan string, 8),
		lastDisplayBuf: make([]rune, bufSize),
		displayBuf:     make([]rune, bufSize),		
	}
}

// AddChar adds a character to the specified line, with overwrite and time-based advancement.
func (tv *TimeVisualiser) AddChar(line int, c rune) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	if line < 0 || line >= tv.numLines {
		return
	}

	now := time.Now()
	for now.Sub(tv.lastAdvance) >= tv.interval {
		nextAdvance := tv.lastAdvance.Add(tv.interval)
		tv.advanceLocked(nextAdvance)
	}

	if tv.cursor < tv.maxChars {
		tv.lines[line][tv.cursor] = c
		tv.notifyLocked()
	}
}

// Advance manually moves the virtual cursor forward.
func (tv *TimeVisualiser) Advance(duration time.Duration) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	for duration >= tv.interval && tv.cursor < tv.maxChars-1 {
		tv.advanceLocked(tv.lastAdvance.Add(tv.interval))
		duration -= tv.interval
	}
}

// Clear resets all lines and the cursor.
func (tv *TimeVisualiser) Clear() {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	for i := range tv.lines {
		for j := range tv.lines[i] {
			tv.lines[i][j] = ' '
		}
	}
	tv.cursor = 0
	tv.lastAdvance = time.Now()
	tv.notifyLocked()
}

// IsFull returns true if the cursor has reached the end of the display.
func (tv *TimeVisualiser) IsFull() bool {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	return tv.cursor >= tv.maxChars-1
}

// Display returns the current display as a string (from preallocated buffer).
func (tv *TimeVisualiser) Display() string {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	return string(tv.buildDisplayLocked())
}

// SetInterval changes the advancement interval.
func (tv *TimeVisualiser) SetInterval(interval time.Duration) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	tv.interval = interval
}

// SetCursor sets the cursor position.
func (tv *TimeVisualiser) SetCursor(pos int) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	if pos >= 0 && pos < tv.maxChars {
		tv.cursor = pos
	}
}

// GetCursor gets the current cursor position.
func (tv *TimeVisualiser) GetCursor() int {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	return tv.cursor
}

// DisplayUpdates provides a channel of display updates.
func (tv *TimeVisualiser) DisplayUpdates() <-chan string {
	return tv.displayUpdates
}

// Internal advance logic — called with lock held.
func (tv *TimeVisualiser) advanceLocked(newAdvanceTime time.Time) {
	tv.lastAdvance = newAdvanceTime
	if tv.cursor < tv.maxChars-1 {
		tv.cursor++
		tv.notifyLocked()
	}
}

// notifyLocked builds and sends display string if changed.
func (tv *TimeVisualiser) notifyLocked() {
	out := tv.buildDisplayLocked()
	if string(out) != string(tv.lastDisplayBuf[:len(out)]) {
		copy(tv.lastDisplayBuf[:], out)
		select {
		case tv.displayUpdates <- string(out):
		default:
		}
	}
}

// buildDisplayLocked creates display output from current lines.
func (tv *TimeVisualiser) buildDisplayLocked() []rune {
	pos := 0
	for i, line := range tv.lines {
		for j := 0; j <= tv.cursor && j < tv.maxChars; j++ {
			tv.displayBuf[pos] = line[j]
			pos++
		}
		if i < tv.numLines-1 {
			tv.displayBuf[pos] = '\n'
			pos++
		}
	}
	return tv.displayBuf[:pos]
}
