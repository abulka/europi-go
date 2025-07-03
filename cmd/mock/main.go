//go:build !tinygo

// Run with go run ./cmd/mock
// Run with go run ./cmd/mock -tea

package main

import (
	"europi/apps"
	hw "europi/controls"
	"europi/display"
	"europi/firmware"
	"europi/logutil"
	"europi/mock"
	"flag"
	"time"
)

var tea = flag.Bool("tea", false, "use Bubble Tea OLED simulation")
var tinyFont = flag.Bool("tinyfont", false, "simulate TinyFont mode (21 chars per line)")

func main() {
	flag.Parse()
	logutil.SetTeaMode(*tea)
	defer logutil.Close()
	time.Sleep(1 * time.Second)
	logutil.Println("Starting...")

	visibleLines := 3
	if *tinyFont {
		visibleLines = 4
	}

	var oled display.IOledDevice
	if *tea {
		oled = display.NewMockOledDeviceTeaWithFont(*tinyFont)
	} else {
		base := display.NewMockOledDeviceWithFont(*tinyFont)
		oled = display.NewBufferedOledDevice(base, visibleLines)
	}
	iox := hw.SetupMockEuroPiWithDisplay(oled)
	if *tea {
		const msg = "EuroPi configured (MOCK TEA ☕️ mode)."
		println(msg)
		logutil.Println(msg)
	} else {
		logutil.Println("EuroPi configured (MOCK 😆 mode).")
	}

	// Register apps
	firmware.RegisterApp(apps.Diagnostic{})
	firmware.RegisterApp(apps.HelloWorld{})
	firmware.RegisterApp(apps.Font8x8{})

	firmware.SplashScreen(iox)
	logutil.Println("Entering main menu loop. Press B2 to select an app, K2 to scroll.")

	// Simulate user input
	go func() {
		mock.SetNumMenuItems(3)
		time.Sleep(1 * time.Second)

		// Visually cycle highlighted menu line: 0 -> 1 -> 2 -> 1 -> 0
		for _, idx := range []int{0, 1, 2, 1, 0} {
			mock.SelectMenuItem(iox.K2, idx)
			time.Sleep(400 * time.Millisecond)
		}

		// Select Diagnostic (index 0)
		mock.SelectMenuItem(iox.K2, 0)
		time.Sleep(300 * time.Millisecond)
		mock.SetButtonPressed(iox.B2, true)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(iox.B2, false)

		// Allow diagnostic app to run for a while - fiddle with some knobs
		mock.SetKnobValue(iox.K2, 10)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(iox.B2, true)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(iox.B2, false)
		time.Sleep(200 * time.Millisecond)
		mock.SetAnalogueInputValue(iox.AIN, 2.5)
		time.Sleep(200 * time.Millisecond)
		mock.SetDigitalInputValue(iox.DIN, true)
		time.Sleep(600 * time.Millisecond)
		mock.SetDigitalInputValue(iox.DIN, false)
		time.Sleep(200 * time.Millisecond)
		mock.SetKnobValue(iox.K2, 0)
		time.Sleep(200 * time.Millisecond)

		mock.ExitToMainMenu(iox)

		// Select HelloWorld (index 1)
		mock.SelectMenuItem(iox.K2, 1)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(iox.B2, true)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(iox.B2, false)

		// Allow other app to run for a while
		time.Sleep(200 * time.Millisecond)

		mock.ExitToMainMenu(iox)

		// Select Font8x8 (index 2)
		mock.SelectMenuItem(iox.K2, 2)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(iox.B2, true)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(iox.B2, false)

		// Allow Font8x8 app to run for a while
		time.Sleep(3 * time.Second)

		logutil.Println("Mock input simulation completed.")
	}()
	for {
		idx := firmware.MenuChooser(iox, visibleLines)
		if idx < 0 {
			logutil.Println("Exiting main menu loop.")
			break
		}
		logutil.Println("Launching app:", firmware.GetAppName(idx))
		firmware.RunApp(idx, iox)
		logutil.Println(firmware.GetAppName(idx), "completed. Returning to menu...")
		firmware.SplashScreen(iox)
	}
}
