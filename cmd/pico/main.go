//go:build tinygo

// tinygo build -target=pico ./cmd/pico
// tinygo flash -target=pico --monitor ./cmd/pico

package main

import (
	hw "europi/controls"
	"europi/firmware"
	"europi/display"
	"europi/apps"
	"time"
)

func main() {
	time.Sleep(1 * time.Second)
	println("Starting...")
	const tinyFont = false
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
	firmware.RegisterApp(apps.Font8x8{})

	firmware.SplashScreen(iox)
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
		firmware.SplashScreen(iox)
	}
}
