package main

import (
	"europi/app"
	"europi/hw"
	"time"
)

const version = "v0.01"

func splashScreen(io *hw.IO) {
	io.Display.ClearDisplay()
	io.Display.WriteLine(0, 10, "EuroPi Simplified")
	io.Display.WriteLine(0, 20, "by TinyGo "+version)
	io.Display.Display()
	time.Sleep(2 * time.Second)
	io.Display.ClearDisplay()
}

func main() {
	time.Sleep(1 * time.Second)
	println("Starting...")

	iox := hw.SetupEuroPi()
	println("EuroPi configured (production mode).")

	// Register apps
	app.RegisterApp(app.Diagnostic{})
	app.RegisterApp(app.HelloWorld{})

	splashScreen(iox)
	println("Entering main menu loop. Press B2 to select an app, K2 to scroll.")

	for {
		idx := app.MenuChooser(iox)
		if idx < 0 {
			println("Exiting main menu loop.")
			break
		}
		println("Launching app:", app.GetAppName(idx))
		app.RunApp(idx, iox)
		println(app.GetAppName(idx), "completed. Returning to menu...")
		splashScreen(iox)
	}
}
