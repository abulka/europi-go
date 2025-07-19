// Real SSD1306 implementation (TinyGo only)
//go:build tinygo

package display

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

type SSD1306Adapter struct {
	dev ssd1306.Device
	// lineYs holds the Y positions for each line in 3-line or 4-line mode, coord is the top of the font, drawn to bottom
	lineYs   []int16
	numLines int // Number of lines (3 or 4)
}

// GetSSD1306 returns the underlying SSD1306 device.
// Its really a *ssd1306.Device 
func (o *SSD1306Adapter) GetSSD1306() any {
	return &o.dev
}

func (o *SSD1306Adapter) ClearDisplay() {
	o.dev.ClearDisplay()
}

func (o *SSD1306Adapter) ClearBuffer() {
	o.dev.ClearBuffer()
}

func (o *SSD1306Adapter) Display() {
	o.dev.Display()
}

// SetNumLines sets the number of lines for the display (3 or 4).
func (o *SSD1306Adapter) SetNumLines(numLines int) {
	if numLines < 3 || numLines > 4 {
		panic("numLines must be 3 or 4")
	}
	if numLines == 3 {
		o.lineYs = []int16{10, 20, 30} // y pixel coordinate of the BOTTOM of each character (for tinyfont mode)
	} else {
		o.lineYs = []int16{7, 15, 23, 31}
	}

	o.numLines = numLines
}

func (o *SSD1306Adapter) NumLines() int {
	return o.numLines
}

/*
Coordinates. The top left corner is (0, 0).
- X ranges from 0 to 127 (left to right)
- Y ranges from 0 to 31 (top to bottom)

- For 8x8font (see other file): you specify the coordinates of the top of the
  font, which is 8 pixels high.
- For tinyfont: you specify the y coordinate of the baseline. This is the
  OPPOSITE of the 8x8font, where you specify the top of the font.

- For pixel operations like display.FillRectangle(x, y, w, h, c): you specify
  the coordinates of the top left corner of the rectangle and the width and
  height of the rectangle.

  NOTES

No, I am not making things up. The issue you described—where text written with
tinyfont.WriteLine does not appear after using a fill rectangle function—is a
real and known behavior with some SSD1306/TinyGo display drivers and the
tinyfont library.

This can happen due to:

Buffer alignment issues (SSD1306 uses 8-pixel-high pages, and clearing with a
rectangle that does not align with the font's baseline or page boundaries can
cause unexpected results). The order of operations: if you clear and then write
to the same area before calling Display(), but the buffer logic is not handling
the overlap as expected, the text may not show. The font rendering may not
overwrite all pixels, especially if the clear rectangle and the font's bounding
box do not match up.

Font Size and Positioning:

Proggy TinySZ8pt7b is a 6x10 font, which means it has a bounding box of 6 pixels wide and 10 pixels high.

https://pkg.go.dev/tinygo.org/x/tinyfont@v0.6.0/proggy

Should only fill three lines in a 32 pixel high display. But if we don't
mind losing the spacing between lines, we can use the full 32 pixels
height for four lines.

*/

func (o *SSD1306Adapter) WriteLine(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(o.lineYs) {
		return
	}
	y := o.lineYs[lineNum]

	// Calculate position for line number (assuming 10px per line)
	// y := int16(lineNum*10 + 10) // +10 for baseline offset

	// 1. Clear old text and any possible highlight area (full width: 128 pixels)
	// ABANDONED because it stops text appearing?  See next diagnostic comment
	// clearY := y - 10
	// clearH := int16(10)
	// fillRectSafe(o.dev, 0, clearY, int16(128), clearH, ColorBlack)
	// println(0, clearY, int16(128), clearH, "line", lineNum, "text", text)

	// 2. test - just clear the top line only each time - WOW this proves that once you clear the top line
	// then try to write to it, no text is written - WEIRD!
	//
	// clearY := int16(0) // Always clear the top line
	// clearH := int16(10) // Height of the line to clear (10 pixels)
	// fillRectSafe(o.dev, 0, clearY, int16(128), clearH, ColorBlack)

	tinyfont.WriteLine(&o.dev, &proggy.TinySZ8pt7b, 0, y, text, ColorWhite)

	// if len(text) >= 2 && text[len(text)-2:] == " *" {
	// 	print("*")
	// } else if text == "" {
	// 	print("_")
	// } else {
	// 	print(".")
	// }
}

func (o *SSD1306Adapter) WriteLineHighlighted(lineNum int, text string) {
	// Highlighted lines are marked with a star at the end
	text += " *"
	o.WriteLine(lineNum, text)
}

// NewOledDeviceTinyFont sets up the I2C and SSD1306 display and returns the display instance.
func NewOledDeviceTinyFont(numLines int) IOledDevice {
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		Frequency: 400000,
		SDA:       machine.GP0,
		SCL:       machine.GP1,
	})
	dev := ssd1306.NewI2C(i2c)
	dev.Configure(ssd1306.Config{
		Address: 0x3C,
		Width:   128,
		Height:  32,
	})
	adapter := &SSD1306Adapter{
		dev: dev,
	}
	adapter.SetNumLines(numLines)
	return adapter
}
