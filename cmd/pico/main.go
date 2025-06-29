//go:build tinygo

package main

import (
	hw "europi/controls"
	"europi/firmware"
	"europi/display"
	"europi/apps"
	"time"
)

const version = "v0.01"

func splashScreen(io *hw.Controls) {
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
	const tinyFont = true
	var oled display.IOledDevice
	if tinyFont {
		println("Using TinyFont for OLED display.")
		oled = display.NewOledDeviceTinyFont()
	} else {
		println("Using 8x8 font for OLED display.")
		oled = display.NewOledDevice8x8()
	}
	iox := hw.SetupEuroPiWithDisplay(oled)
	println("EuroPi configured (production mode).")

	// Register apps
	firmware.RegisterApp(apps.Diagnostic{})
	firmware.RegisterApp(apps.HelloWorld{})

	splashScreen(iox)
	println("Entering main menu loop. Press B2 to select an app, K2 to scroll.")

	for {
		idx := firmware.MenuChooser(iox)
		if idx < 0 {
			println("Exiting main menu loop.")
			break
		}
		println("Launching app:", firmware.GetAppName(idx))
		firmware.RunApp(idx, iox)
		println(firmware.GetAppName(idx), "completed. Returning to menu...")
		splashScreen(iox)
	}
}
