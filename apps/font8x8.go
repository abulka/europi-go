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

// When in 8x8 font mode, only 16 characters per line, cos 16 x 8 = 128 pixels - see up to and including "P"
//   ABCDEFGHIJKLMNOP|QRSTUVWXYZ
//   abcdefghijklmnop|qrstuvwxyz
//   0123456789!@#$%^|&*()./-_=+
// When in TinyFont mode, 21 characters per line - see up to and including "U"
//   ABCDEFGHIJKLMNOPQRSTU|VWXYZ
//   abcdefghijklmnopqrstu|vwxyz
//   0123456789!@#$%^&*().|/-_=+

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
