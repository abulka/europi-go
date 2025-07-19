// HelloWorld app
package apps

import (
	"europi/controls"
	"europi/firmware"
	"europi/logutil"
	"time"
)

type HelloWorld struct{}

func (c HelloWorld) Name() string { return "Hello World" }

func (c HelloWorld) Run(hw *controls.Controls) {
	logutil.Println("Hello, World!")
	hw.Display.ClearDisplay()
	hw.Display.WriteLine(0, "Hello, World!")
	hw.Display.Display()

	for {
		if firmware.ShouldExit(hw) {
			break
		}
	}
	logutil.Println("Exiting HelloWorld application.")
	hw.Display.ClearDisplay()
	hw.Display.Display()
	logutil.Println("Display cleared. Goodbye!")
	time.Sleep(1 * time.Second)
}
