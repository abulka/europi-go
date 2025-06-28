// App interface, registry, shared app logic
package firmware

import (
	hw "europi/controls"
	"time"
)

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

// MenuDisplayState holds the current state of the menu display to avoid unnecessary redraws and flicker.
type MenuDisplayState struct {
	Lines [3]string
}

func (m *MenuDisplayState) IsDirty(lines [3]string) bool {
	for i := 0; i < 3; i++ {
		if m.Lines[i] != lines[i] {
			return true
		}
	}
	return false
}

func (m *MenuDisplayState) Update(lines [3]string) {
	copy(m.Lines[:], lines[:])
}

// MenuChooser displays a scrollable menu of registered apps, allows selection with K2, launch with B2
func MenuChooser(io *hw.Controls) int {
	numApps := len(appRegistry)
	if numApps == 0 {
		return -1
	}
	selected := 0
	lastK2 := io.K2.Value()
	const debounceMs = 150
	lastAction := time.Now()
	displayState := &MenuDisplayState{}
	for {
		k2 := io.K2.Value()
		if abs(k2-lastK2) > 2 && time.Since(lastAction) > debounceMs*time.Millisecond {
			if k2 > lastK2 {
				selected++
			} else if k2 < lastK2 {
				selected--
			}
			if selected < 0 {
				selected = 0
			}
			if selected >= numApps {
				selected = numApps - 1
			}
			lastK2 = k2
			lastAction = time.Now()
		}
		start := selected - 1
		if start < 0 {
			start = 0
		}
		if start > numApps-3 {
			start = numApps - 3
		}
		if start < 0 {
			start = 0
		}
		var lines [3]string
		for i := 0; i < 3 && (start+i) < numApps; i++ {
			idx := start + i
			name := appRegistry[idx].Name()
			line := name
			if idx == selected {
				line += " *"
			}
			lines[i] = line
		}
		if displayState.IsDirty(lines) {
			io.Display.ClearDisplay()
			for i := 0; i < 3; i++ {
				if lines[i] != "" {
					io.Display.WriteLine(0, int16(10+10*i), lines[i])
				}
			}
			io.Display.Display()
			displayState.Update(lines)
		}
		if io.B2.Pressed() && !io.B1.Pressed() {
			for io.B2.Pressed() {
				time.Sleep(10 * time.Millisecond)
			}
			return selected
		}
		if ShouldExit(io) {
			return -1
		}
		time.Sleep(30 * time.Millisecond)
	}
}

// abs helper for menu logic
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
