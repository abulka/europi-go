package firmware

import (
	hw "europi/controls"
	"time"
)

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
		for i := 0; i < 3 && (start+i) < numApps; i++ {
			idx := start + i
			name := appRegistry[idx].Name()
			io.Display.WriteLine(0, int16(10+10*i), name)
		}
		io.Display.HighlightLn(selected - start)
		io.Display.Display()
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
