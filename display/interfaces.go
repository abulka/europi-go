// IOledDevice interface, shared display logic
package display

import "image/color"

type IOledDevice interface {
	// ClearDisplay clears the display content.
	ClearDisplay()
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
