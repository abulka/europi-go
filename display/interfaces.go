// IOledDevice interface, shared display logic
package display

type IOledDevice interface {
	// ClearDisplay clears the display content.
	ClearDisplay()
	// Display updates the display with the current content.
	Display()
	// WriteLine writes a line of text to the display at the specified line number.
	WriteLine(lineNum int, text string)
	WriteLineHighlighted(lineNum int, text string)
}
