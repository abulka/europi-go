//go:build tinygo

// Building only
// tinygo build -target=pico ./cmd/pico
// or for TinyFont support:
// tinygo build -tags tinyfont -target=pico ./cmd/

// Flashing to Raspberry Pi Pico EuroPi
// tinygo flash -target=pico --monitor ./cmd/pico
// or for TinyFont support:
// tinygo flash -tags tinyfont -target=pico --monitor ./cmd/pico

package main

import (
	"europi/apps"
	hw "europi/controls"
	"europi/display"
	"europi/firmware"
	"time"
)

func main() {
	time.Sleep(1 * time.Second)
	println("Starting...")
	var oled display.IOledDevice
	if tinyFont {
		println("Using TinyFont for OLED display.")
		oled = display.NewOledDeviceTinyFont()
	} else {
		println("Using 8x8 font for OLED display.")
		oled = display.NewOledDevice8x8()
	}
	
	// wrap with buffered display decorator
	oled = display.NewBufferedDisplayWithFont(oled, tinyFont)

	iox := hw.SetupEuroPiWithDisplay(oled)
	println("EuroPi configured (production mode).")

	// Register apps
	firmware.RegisterApp(apps.Diagnostic{})
	firmware.RegisterApp(apps.HelloWorld{})
	firmware.RegisterApp(apps.Font8x8{})

	firmware.SplashScreen(iox)
	println("Entering main menu loop. Press B2 to select an app, K2 to scroll.")

	visibleLines := 3
	if tinyFont {
		visibleLines = 4
	}
	for {
		idx := firmware.MenuChooser(iox, visibleLines)
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
