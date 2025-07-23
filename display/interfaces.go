// IOledDevice interface, shared display logic
package display

import "image/color"

type IOledDevice interface {
	// 3 or 4 lines for OLED display
	NumLines() int
	SetNumLines(n int)
	// Returns the underlying SSD1306 device if available, otherwise nil (for mocks)
	GetSSD1306() any
	// ClearDisplay clears the display content.
	ClearDisplay()
	ClearBuffer()
	// Display updates the display with the current content.
	Display()
	// WriteLine writes a line of text to the display at the specified line number.
	WriteLine(lineNum int, text string)
	WriteLineHighlighted(lineNum int, text string)
}

// Common color constants for OLED rendering
var (
	ColorBlack = color.RGBA{0, 0, 0, 255}
	ColorWhite = color.RGBA{255, 255, 255, 255}
)

// The raw device not my higher level IOledDevice
type ISSD1306Device interface {
	SetPixel(x, y int16, c color.RGBA)
	Display() error
	ClearDisplay()
	ClearBuffer()
	FillRectangle(x, y, width, height int16, c color.RGBA) error
}