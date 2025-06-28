// HelloWorld app
package app

import (
	hw "europi/controls"
	"time"
)

type HelloWorld struct{}

func (c HelloWorld) Name() string { return "Hello World" }

func (c HelloWorld) Run(io *hw.Controls) {
	println("Hello, World!")
	io.Display.ClearDisplay()
	io.Display.WriteLine(0, 10, "Hello, World!")
	io.Display.Display()

	for {
		if ShouldExit(io) {
			break
		}
	}
	println("Exiting HelloWorld application.")
	io.Display.ClearDisplay()
	io.Display.Display()
	println("Display cleared. Goodbye!")
	time.Sleep(1 * time.Second)
}
