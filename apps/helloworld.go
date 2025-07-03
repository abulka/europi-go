// HelloWorld app
package apps

import (
	hw "europi/controls"
	"europi/firmware"
	"europi/logutil"
	"time"
)

type HelloWorld struct{}

func (c HelloWorld) Name() string { return "Hello World" }

func (c HelloWorld) Run(io *hw.Controls) {
	logutil.Println("Hello, World!")
	io.Display.ClearDisplay()
	io.Display.WriteLine(0, "Hello, World!")
	io.Display.Display()

	for {
		if firmware.ShouldExit(io) {
			break
		}
	}
	logutil.Println("Exiting HelloWorld application.")
	io.Display.ClearDisplay()
	io.Display.Display()
	logutil.Println("Display cleared. Goodbye!")
	time.Sleep(1 * time.Second)
}
