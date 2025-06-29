//go:build !tinygo

package main

import (
	"europi/apps"
	hw "europi/controls"
	"europi/display"
	"europi/firmware"
	"europi/logutil"
	"flag"
	"time"
)

const version = "v0.01"

var tea = flag.Bool("tea", false, "use Bubble Tea OLED simulation")

func splashScreen(io *hw.Controls) {
	io.Display.ClearDisplay()
	io.Display.WriteLine(0, 10, "EuroPi Simplified")
	io.Display.WriteLine(0, 20, "by TinyGo "+version)
	io.Display.Display()
	time.Sleep(2 * time.Second)
	io.Display.ClearDisplay()
}

// Helper functions for mock input, since SetValue and SetPressed do not exist on the IKnob and IButton interfaces.
// But they do exist on the mock implementations used in tests.
func SetKnobValue(knob hw.IKnob, v int) {
	if mock, ok := knob.(interface{ SetValue(int) }); ok {
		mock.SetValue(v)
	}
}

func SetButtonPressed(btn hw.IButton, pressed bool) {
	if mock, ok := btn.(interface{ SetPressed(bool) }); ok {
		mock.SetPressed(pressed)
	}
}

func main() {
	flag.Parse()
	logutil.SetTeaMode(*tea)
	defer logutil.Close()
	time.Sleep(1 * time.Second)
	logutil.Println("Starting...")

	var oled display.IOledDevice
	if *tea {
		oled = display.NewMockOledDeviceTea()
	} else {
		oled = display.NewMockOledDevice()
	}
	iox := hw.SetupMockEuroPiWithDisplay(oled)
	logutil.Println("EuroPi configured (MOCK 😆 mode).")

	// Register apps
	firmware.RegisterApp(apps.Diagnostic{})
	firmware.RegisterApp(apps.HelloWorld{})

	splashScreen(iox)
	logutil.Println("Entering main menu loop. Press B2 to select an app, K2 to scroll.")

	// Simulate user input
	go func() {
		time.Sleep(1 * time.Second)
		SetKnobValue(iox.K2, 5)
		time.Sleep(300 * time.Millisecond)
		SetKnobValue(iox.K2, 10)
		time.Sleep(300 * time.Millisecond)
		SetKnobValue(iox.K2, 0)
		time.Sleep(300 * time.Millisecond)

		// Simulate pressing B2 to select an app
		SetButtonPressed(iox.B2, true)
		time.Sleep(200 * time.Millisecond)
		SetButtonPressed(iox.B2, false)

		// Allow diagnostic app to run for a while
		time.Sleep(1 * time.Second)

		// Exit the app after 5 seconds by pressing B1 and B2 simultaneously
		SetButtonPressed(iox.B1, true)
		SetButtonPressed(iox.B2, true)
		time.Sleep(3 * time.Second) // Simulate holding both buttons for > 2 seconds to exit
		SetButtonPressed(iox.B1, false)
		SetButtonPressed(iox.B2, false)
		time.Sleep(2 * time.Second) // Wait for splash screen to clear

		// Simulate scrolling to different app and selecting it
		SetKnobValue(iox.K2, 10)
		time.Sleep(200 * time.Millisecond)
		SetButtonPressed(iox.B2, true)
		time.Sleep(200 * time.Millisecond)
		SetButtonPressed(iox.B2, false)

		// Allow other app to run for a while
		time.Sleep(2 * time.Second)

		logutil.Println("Mock input simulation completed.")
	}()
	for {
		idx := firmware.MenuChooser(iox)
		if idx < 0 {
			logutil.Println("Exiting main menu loop.")
			break
		}
		logutil.Println("Launching app:", firmware.GetAppName(idx))
		firmware.RunApp(idx, iox)
		logutil.Println(firmware.GetAppName(idx), "completed. Returning to menu...")
		splashScreen(iox)
	}
}
