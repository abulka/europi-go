//go:build !tinygo

// Run with go run ./cmd/mock
// Run with go run ./cmd/mock -tinyfont
// Run with go run ./cmd/mock -tea
// Run with go run ./cmd/mock -tea -tinyfont

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

	// Always use buffered display for all mock modes (Tea or not)
	buffered := true // Set to false to disable buffering for all modes

	var oled display.IOledDevice
	if *tea {
		base := display.NewMockOledDeviceTeaWithFont(*tinyFont)
		if buffered {
			oled = display.NewBufferedDisplayWithFont(base, *tinyFont)
		} else {
			oled = base
		}
	} else {
		base := display.NewMockOledDeviceWithFont(*tinyFont)
		if buffered {
			oled = display.NewBufferedDisplayWithFont(base, *tinyFont)
		} else {
			oled = base
		}
	}
	iox := hw.SetupMockEuroPiWithDisplay(oled)
	if *tea {
		mode := "MOCK TEA ☕️ mode"
		if buffered {
			mode += ", buffered"
		}
		msg := "EuroPi configured (" + mode + ")."
		println(msg)
		logutil.Println(msg)
	} else {
		mode := "MOCK 😆 mode"
		if buffered {
			mode += ", buffered"
		}
		logutil.Println("EuroPi configured (" + mode + ").")
	}

	// Register apps
	firmware.RegisterApp(apps.Diagnostic{})
	firmware.RegisterApp(apps.HelloWorld{})
	firmware.RegisterApp(apps.Font8x8{})

	firmware.SplashScreen(iox)
	logutil.Println("Entering main menu loop. Press B2 to select an app, K2 to scroll.")

	// Simulate user input
	go func() {
		numMenuItems := firmware.NumRegisteredApps()
		if numMenuItems == 0 {
			numMenuItems = 1
		}
		mock.SetNumMenuItems(numMenuItems)
		time.Sleep(100 * time.Millisecond)

		// Visually cycle highlighted menu line: 0 -> 1 -> ... -> n-1 -> ... -> 0
		cycleThroughMenuItems(numMenuItems, iox)

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

func cycleThroughMenuItems(numMenuItems int, iox *hw.Controls) {
	var cycle []int
	for i := 0; i < numMenuItems; i++ {
		cycle = append(cycle, i)
	}
	for i := numMenuItems - 2; i > 0; i-- {
		cycle = append(cycle, i)
	}
	for _, idx := range cycle {
		mock.SelectMenuItem(iox.K2, idx)
		time.Sleep(400 * time.Millisecond)
	}
}
