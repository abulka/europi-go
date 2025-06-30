// Font8x8 app
package apps

import (
	hw "europi/controls"
	"europi/firmware"
	"europi/logutil"
	"time"
)

type Font8x8 struct{}

func (c Font8x8) Name() string { return "Font8x8" }

func (c Font8x8) Run(io *hw.Controls) {
	io.Display.ClearDisplay()
	io.Display.WriteLine(0, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	io.Display.WriteLine(0, 20, "abcdefghijklmnopqrstuvwxyz")
	io.Display.WriteLine(0, 30, "0123456789!@#$%^&*()./-_=+")
	io.Display.Display()

	for {
		if firmware.ShouldExit(io) {
			break
		}
	}
	logutil.Println("Exiting Font8x8 application.")
	io.Display.ClearDisplay()
	io.Display.Display()
	logutil.Println("Display cleared. Goodbye!")
	time.Sleep(1 * time.Second)
}
