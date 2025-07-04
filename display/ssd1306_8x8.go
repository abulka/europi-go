//go:build tinygo

package display

import (
	"image/color"
	"machine"
	"tinygo.org/x/drivers/ssd1306"
)

type SSD1306Adapter8x8 struct {
	dev          ssd1306.Device
	fontRenderer *Font8x8Renderer
}

func (o *SSD1306Adapter8x8) ClearDisplay() {
	o.dev.ClearDisplay()
}

func (o *SSD1306Adapter8x8) Display() {
	o.dev.Display()
}

func (o *SSD1306Adapter8x8) WriteLine(lineNum int, text string) {
	// Calculate position for line number (8px per line for 8x8 font)
	y := int16(lineNum * 8)
	o.fontRenderer.WriteLine(o.dev, 0, y, text, color.RGBA{255, 255, 255, 255})
}

func (o *SSD1306Adapter8x8) WriteLineHighlighted(lineNum int, text string) {
	// Highlight this line (reverse video)
	y := int16(lineNum * 8)
	o.fontRenderer.HighlightLine(o.dev, 0, y, text, 1, color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255})
}

// NewOledDevice8x8 creates a new SSD1306 device with 8x8 font support
func NewOledDevice8x8() IOledDevice {
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
	return &SSD1306Adapter8x8{
		dev:          dev,
		fontRenderer: NewFont8x8Renderer(),
	}
}
