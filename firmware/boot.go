// App interface, registry, shared app logic
package firmware

import (
	"europi/controls"
	"time"
)

const version = "v0.01"

type App interface {
	Name() string
	Run(hw *controls.Controls)
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

func NumRegisteredApps() int {
	return len(appRegistry)
}

func RunApp(idx int, hw *controls.Controls) {
	if idx >= 0 && idx < len(appRegistry) {
		appRegistry[idx].Run(hw)
	}
}

// ShouldExit returns true if both B1 and B2 are pressed and held for 2s
var doubleButtonPressLastMs int64 = 0

func ShouldExit(hw *controls.Controls) bool {
	now := time.Now().UnixMilli()
	if hw.B1.Pressed() && hw.B2.Pressed() {
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

func SplashScreen(hw *controls.Controls) {
	hw.Display.ClearDisplay()
	hw.Display.WriteLine(0, "EuroPi Simplified")
	hw.Display.WriteLine(1, "by TinyGo "+version)
	hw.Display.Display()
	time.Sleep(1 * time.Second)
	hw.Display.ClearDisplay()
}
