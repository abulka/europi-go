//go:build tinygo

package display

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

type SSD1306Adapter8x8 struct {
	// dev is the underlying SSD1306 device, x ranges from 0 to 127 (left to right), y ranges from 0 to 31 (top to bottom)
	dev ssd1306.Device
	// lineYs holds the Y positions for each line in 3-line or 4-line mode, coord is the top of the font, drawn to bottom
	lineYs []int16
	// Highlight margins (in pixels)
	HighlightMarginTop    int16
	HighlightMarginBottom int16
}

func (o *SSD1306Adapter8x8) ClearDisplay() {
	o.dev.ClearDisplay()
}

func (o *SSD1306Adapter8x8) Display() {
	o.dev.Display()
}

// NewOledDevice8x8 creates a new SSD1306 device with 8x8 font support
// Pass numLines = 3 or 4
func NewOledDevice8x8(numLines int) IOledDevice {
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
	adapter := &SSD1306Adapter8x8{
		dev: dev,
	}
	adapter.SetLineMode(numLines)
	return adapter
}

// SetLineMode switches between 3-line and 4-line modes and sets highlight margins
func (o *SSD1306Adapter8x8) SetLineMode(numLines int) {
	if numLines == 3 {
		o.lineYs = []int16{2, 12, 22}
		o.HighlightMarginTop = 1
		o.HighlightMarginBottom = 1
	} else {
		o.lineYs = []int16{0, 8, 16, 24}
		o.HighlightMarginTop = 0
		o.HighlightMarginBottom = 0
	}
}

func (o *SSD1306Adapter8x8) WriteLine(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(o.lineYs) {
		return
	}
	y := o.lineYs[lineNum]

	// Clear old text and any possible highlight area (full width: 128 pixels)
	clearY := y - o.HighlightMarginTop
	clearH := int16(8) + o.HighlightMarginTop + o.HighlightMarginBottom
	fillRectSafe(o.dev, 0, clearY, int16(128), clearH, ColorBlack)

	DrawFont8x8Text(o.dev, 0, y, text, ColorWhite)
}

func (o *SSD1306Adapter8x8) WriteLineHighlighted(lineNum int, text string) {
	if lineNum < 0 || lineNum >= len(o.lineYs) {
		return
	}
	y := o.lineYs[lineNum]
	textW := int16(len(text) * 8)
	if textW > 128 {
		textW = 128
	}
	textH := int16(8)
	rectX := int16(0)
	rectY := y - o.HighlightMarginTop
	rectW := textW
	rectH := textH + o.HighlightMarginTop + o.HighlightMarginBottom
	fillRectSafe(o.dev, rectX, rectY, rectW, rectH, ColorWhite)
	DrawFont8x8Text(o.dev, 0, y, text, ColorBlack)
	println("Line", lineNum, "at y:", y, "highlight:", rectX, rectY, rectW, rectH, "text:", text)
}

// fillRectSafe clamps the rectangle to the 128x32 display area. Why: If you
// attempt to draw a rectangle that has any pixel off-screen,
// display.FillRectangle does nothing, so we clamp all values to ensure
// something is always drawn.
func fillRectSafe(display ssd1306.Device, x, y, w, h int16, c color.RGBA) {
	origX, origY, origW, origH := x, y, w, h
	clamped := false
	debug := false

	// Clamp x and y to display bounds
	if x < 0 {
		w += x // reduce width by how much x is negative
		x = 0
		clamped = true
	}
	if y < 0 {
		h += y // reduce height by how much y is negative
		y = 0
		clamped = true
	}
	// Clamp width and height so rectangle stays within display
	if x+w > 128 {
		w = 128 - x
		clamped = true
	}
	if y+h > 32 {
		h = 32 - y
		clamped = true
	}
	// If rectangle is completely off-screen, do nothing
	if w <= 0 || h <= 0 && debug {
		println("fillRectSafe: Rectangle completely off-screen, nothing to draw. Original:", origX, origY, origW, origH)
		return
	}
	if clamped && debug {
		println("fillRectSafe: Clamped rectangle from", origX, origY, origW, origH, "to", x, y, w, h)
	}
	display.FillRectangle(x, y, w, h, c)
}
