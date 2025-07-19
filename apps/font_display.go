// Font8x8 app
package apps

import (
	"europi/controls"
	"europi/firmware"
	"europi/logutil"
	"time"
)

type FontDisplay struct{}

func (c FontDisplay) Name() string { return "Font" }

// When in 8x8 font mode, only 16 characters per line, cos 16 x 8 = 128 pixels - see up to and including "P"
//   ABCDEFGHIJKLMNOP|QRSTUVWXYZ
//   abcdefghijklmnop|qrstuvwxyz
//   0123456789!@#$%^|&*()./-_=+
//   xyzABCDEFGHIJKLMNOP...
//
// When in TinyFont mode, 21 characters per line - see up to and including "U"
//   ABCDEFGHIJKLMNOPQRSTU|VWXYZ
//   abcdefghijklmnopqrstu|vwxyz
//   0123456789!@#$%^&*().|/-_=+
//   xyzABCDEFGHIJKLMNOP...

func (c FontDisplay) Run(hw *controls.Controls) {
	hw.Display.ClearBuffer()
	hw.Display.WriteLine(0, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hw.Display.WriteLine(1, "abcdefghijklmnopqrstuvwxyz")
	hw.Display.WriteLine(2, "0123456789!@#$%^&*()./-_=+")
	hw.Display.WriteLine(3, "EuroPi-Go Font Display")
	hw.Display.Display()

	for {
		if firmware.ShouldExit(hw) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	logutil.Println("Exiting Font application.")
	hw.Display.ClearBuffer()
	hw.Display.Display()
	logutil.Println("Display cleared. Goodbye!")
	time.Sleep(1 * time.Second)
}
