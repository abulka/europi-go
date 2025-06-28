// IOledDevice interface, shared display logic
package display

// IOledDevice interface defines the methods for interacting with an OLED display.
type IOledDevice interface {
	// ClearDisplay clears the display content.
	ClearDisplay()
	// Display updates the display with the current content.
	Display()
	// WriteLine writes a line of text to the display at the specified coordinates.
	WriteLine(x, y int16, text string)
}
