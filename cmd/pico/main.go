//go:build tinygo

// Flashing to Raspberry Pi Pico EuroPi
// For building only, replace `flash` with `build` in the commands below.
// tinygo flash -target=pico --monitor ./cmd/pico
// tinygo flash -tags lotslines -target=pico --monitor ./cmd/pico
// tinygo flash -tags tinyfont -target=pico --monitor ./cmd/pico
// tinygo flash -tags lotslines,tinyfont -target=pico --monitor ./cmd/pico

package main

import (
	"europi/apps"
	"europi/apps/pixel_animation"
	"europi/cmd/pico/config"
	"europi/controls"
	"europi/display"
	"europi/firmware"
	"time"
)

func main() {
	time.Sleep(1 * time.Second)
	println("Starting...")
	var oled display.IOledDevice
	if config.TinyFont {
		println("Using TinyFont for OLED display.")
		oled = display.NewOledDeviceTinyFont(config.NumLines)
	} else {
		println("Using 8x8 font for OLED display.")
		oled = display.NewOledDevice8x8(config.NumLines)
	}

	// Wrap with buffered display decorator (optional)
	// Not needed anymore now that we know to use ClearBuffer() calls.
	// But it DID reduce the amount of calls to backend dev.Display() for the menuchooser when it was coded to call Display() after every K2 knob change. Now its smarter.
	// oled = display.NewBufferedDisplay(oled, config.NumLines)

	hw := controls.SetupEuroPiWithDisplay(oled)
	println("EuroPi configured (production mode).")

	// Register apps
	firmware.RegisterApp(apps.TriggerGateDelay{})
	firmware.RegisterApp(apps.TriggerGateDelay2{})
	firmware.RegisterApp(apps.TriggerMirror{})
	firmware.RegisterApp(apps.Diagnostic{})
	firmware.RegisterApp(apps.HelloWorld{})
	firmware.RegisterApp(apps.FontDisplay{})
	firmware.RegisterApp(apps.MenuFun{})
	firmware.RegisterApp(pixel_animation.Pixels{})
	firmware.RegisterApp(pixel_animation.Pixels2{})
	firmware.RegisterApp(pixel_animation.Pixels3{})

	firmware.SplashScreen(hw)
	println("Entering main menu loop. Press B2 to select an app, K2 to scroll.")

	visibleLines := config.NumLines
	for {
		hw.Display.SetNumLines(visibleLines) // Ensure display is set to the correct number of lines just in case an app changed it
		idx := firmware.MenuChooser(hw, visibleLines)
		if idx < 0 {
			println("Cannot Exit main menu loop.")
			continue
		}
		println("Launching app:", firmware.GetAppName(idx))
		firmware.RunApp(idx, hw)
		println(firmware.GetAppName(idx), "completed. Returning to menu...")
		firmware.SplashScreen(hw)
	}
}
