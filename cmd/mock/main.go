//go:build !tinygo

// Run with go run ./cmd/mock
// Run with go run ./cmd/mock -tea -tinyfont -lotslines

package main

import (
	"europi/apps"
	"europi/controls"
	"europi/display"
	"europi/firmware"
	"europi/logutil"
	"europi/mock"
	"flag"
	"strconv"
	"time"
)

var tea = flag.Bool("tea", false, "use Bubble Tea OLED simulation")
var tinyFont = flag.Bool("tinyfont", false, "simulate TinyFont mode (21 chars per line)")
var lotsLines = flag.Bool("lotslines", false, "simulate 4 lines of text (default is 3 lines)")

func main() {
	flag.Parse()
	logutil.SetTeaMode(*tea)
	defer logutil.Close()
	time.Sleep(1 * time.Second)
	logutil.Println("Starting...")

	numLines := 3
    if *lotsLines {
        numLines = 4
    }
	lineLen := 16
	if *tinyFont {
		lineLen = 21 // TinyFont has 21 chars per line
	}

	// Always use buffered display for all mock modes (Tea or not)
	buffered := false // Set to false to disable buffering for all modes

	var oled display.IOledDevice
	if *tea {
		oled = display.NewMockOledDeviceTea(numLines, lineLen)
	} else {
		oled = display.NewMockOledDevice(numLines, lineLen)
	}
	if buffered {
		oled = display.NewBufferedDisplay(oled, numLines)
	}
	hw := controls.SetupMockEuroPiWithDisplay(oled)
	mode := "MOCK "
	if *tea {
		mode += "TEA ☕️ "
	} else {
		mode += "😆 "
	}
	if buffered {
		mode += ", buffered"
	}
	msg := "EuroPi configured (" + mode + ").NumLines: " + strconv.Itoa(hw.Display.NumLines())
	logutil.Println(msg)

	// Register apps
	firmware.RegisterApp(apps.Diagnostic{})
	firmware.RegisterApp(apps.HelloWorld{})
	firmware.RegisterApp(apps.FontDisplay{})
	firmware.RegisterApp(apps.MenuFun{})

	firmware.SplashScreen(hw)
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
		// cycleThroughMenuItems(numMenuItems, hw)
		// time.Sleep(1 * time.Second)

		// Select menu fun app (index 3)
		mock.SelectMenuItem(hw.K2, 0)
		time.Sleep(1 * time.Second)
		mock.SelectMenuItem(hw.K2, 1) 
		time.Sleep(1 * time.Second)
		mock.SelectMenuItem(hw.K2, 2)
		time.Sleep(1 * time.Second)
		mock.SelectMenuItem(hw.K2, 3)
		time.Sleep(2 * time.Second)
		
		// Press B2 to select the app
		mock.ButtonPress(hw.B2)
		time.Sleep(2 * time.Second)

		// MenuFun app
		mock.SelectMenuItem(hw.K2, 0) // Select first item
		time.Sleep(2 * time.Second)
		mock.ButtonPress(hw.B2)
		time.Sleep(1 * time.Second)
		mock.ButtonPress(hw.B1)
		time.Sleep(1 * time.Second)

		mock.SelectMenuItem(hw.K2, 1) // Select another item
		time.Sleep(2 * time.Second)
		mock.ButtonPress(hw.B2)
		time.Sleep(1 * time.Second)
		mock.ButtonPress(hw.B1)
		time.Sleep(1 * time.Second)

		mock.ExitToMainMenu(hw)
		logutil.Println("Returning to main menu...")

		// Select Diagnostic app (index 0)
		mock.SelectMenuItem(hw.K2, 0)
		time.Sleep(2 * time.Second)
		mock.ButtonPress(hw.B2)
		time.Sleep(1 * time.Second)

		// Allow diagnostic App to run for a while - fiddle with some knobs
		mock.SetKnobValue(hw.K2, 10)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(hw.B2, true)
		time.Sleep(200 * time.Millisecond)
		mock.SetButtonPressed(hw.B2, false)
		time.Sleep(200 * time.Millisecond)
		mock.SetAnalogueInputValue(hw.AIN, 2.5)
		time.Sleep(200 * time.Millisecond)
		mock.SetDigitalInputValue(hw.DIN, true)
		time.Sleep(600 * time.Millisecond)
		mock.SetDigitalInputValue(hw.DIN, false)
		time.Sleep(200 * time.Millisecond)
		mock.SetKnobValue(hw.K2, 0)
		time.Sleep(200 * time.Millisecond)

		mock.ExitToMainMenu(hw)

		// Select HelloWorld App (index 1)
		mock.SelectMenuItem(hw.K2, 1)
		time.Sleep(2 * time.Second)
		mock.ButtonPress(hw.B2)
		time.Sleep(1 * time.Second)

		mock.ExitToMainMenu(hw)

		// Select Font App (index 2)
		mock.SelectMenuItem(hw.K2, 2)
		time.Sleep(2 * time.Second)
		mock.ButtonPress(hw.B2)
		time.Sleep(1 * time.Second)

		mock.ExitToMainMenu(hw)

		logutil.Println("Mock input simulation completed.")
	}()
	for {
		idx := firmware.MenuChooser(hw, numLines)
		if idx < 0 {
			logutil.Println("Exiting main menu loop.")
			break
		}
		logutil.Println("Launching app:", firmware.GetAppName(idx))
		firmware.RunApp(idx, hw)
		logutil.Println(firmware.GetAppName(idx), "completed. Returning to menu...")
		firmware.SplashScreen(hw)
	}
}

func cycleThroughMenuItems(numMenuItems int, hw *controls.Controls) {
	var cycle []int
	for i := 0; i < numMenuItems; i++ {
		cycle = append(cycle, i)
	}
	for i := numMenuItems - 2; i >= 0; i-- {
		cycle = append(cycle, i)
	}
	for _, idx := range cycle {
		mock.SelectMenuItem(hw.K2, idx)
		time.Sleep(400 * time.Millisecond)
	}
}
