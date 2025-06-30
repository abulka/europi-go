// App interface, registry, shared app logic
package firmware

import (
	hw "europi/controls"
	"time"
)

const version = "v0.01"

type App interface {
	Name() string
	Run(io *hw.Controls)
}

var appRegistry []App

func RegisterApp(app App) {
	appRegistry = append(appRegistry, app)
}

func GetAppName(idx int) string {
	if idx >= 0 && idx < len(appRegistry) {
		return appRegistry[idx].Name()
	}
	return ""
}

func RunApp(idx int, io *hw.Controls) {
	if idx >= 0 && idx < len(appRegistry) {
		appRegistry[idx].Run(io)
	}
}


// ShouldExit returns true if both B1 and B2 are pressed and held for 2s
var doubleButtonPressLastMs int64 = 0

func ShouldExit(io *hw.Controls) bool {
	now := time.Now().UnixMilli()
	if io.B1.Pressed() && io.B2.Pressed() {
		if doubleButtonPressLastMs == 0 {
			doubleButtonPressLastMs = now
		} else if now-doubleButtonPressLastMs >= 2000 {
			doubleButtonPressLastMs = 0
			return true
		}
	} else {
		doubleButtonPressLastMs = 0
	}
	return false
}

func SplashScreen(io *hw.Controls) {
	io.Display.ClearDisplay()
	io.Display.WriteLine(0, 10, "EuroPi Simplified")
	io.Display.WriteLine(0, 20, "by TinyGo "+version)
	io.Display.Display()
	time.Sleep(2 * time.Second)
	io.Display.ClearDisplay()
}
