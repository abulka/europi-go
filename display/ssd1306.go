// Real SSD1306 implementation (TinyGo only)
//go:build tinygo

package display

import (
	"image/color"

	"machine"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

type SSD1306Adapter struct {
	dev ssd1306.Device
}

func (o *SSD1306Adapter) ClearDisplay() {
	o.dev.ClearDisplay()
}

func (o *SSD1306Adapter) Display() {
	o.dev.Display()
}

func (o *SSD1306Adapter) WriteLine(x, y int16, text string) {
	tinyfont.WriteLine(&o.dev, &proggy.TinySZ8pt7b, x, y, text, color.RGBA{255, 255, 255, 255})
}

// InitDisplay sets up the I2C and SSD1306 display and returns the display instance.
func NewOledDeviceTinyFont() IOledDevice {
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
	return &SSD1306Adapter{dev: dev}
}
