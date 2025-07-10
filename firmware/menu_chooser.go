package firmware

import (
	hw "europi/controls"
	"time"
)

// MenuChooser displays a scrollable menu of registered apps, allows selection with K2, launch with B2
func MenuChooser(io *hw.Controls, visibleLines int) int {
	numApps := len(appRegistry)
	if numApps == 0 {
		return -1
	}
	names := make([]string, numApps)
	for i, app := range appRegistry {
		names[i] = app.Name()
	}
	return ScrollingMenu(names, io, visibleLines)
}

// ScrollingMenu displays a scrollable menu of items, allows selection with K2, launch with B2
// Returns the selected index, or -1 if exited
func ScrollingMenu(items []string, io *hw.Controls, visibleLines int) int {
	numItems := len(items)
	if numItems == 0 {
		return -1
	}
	selected := 0
	lastK2 := -1 // Initialize to an invalid value to force display update on first loop
	for {
		k2 := io.K2.Value()
		updateDisplay := false
		if k2 != lastK2 {
			selected = (k2 * numItems) / 100
			if selected < 0 {
				selected = 0
			}
			if selected >= numItems {
				selected = numItems - 1
			}
			lastK2 = k2
			updateDisplay = true
		}
		// Only update the display if K2 changed, or on the first loop
		if updateDisplay {
			io.Display.ClearDisplay()
			if numItems <= visibleLines {
				// No windowing, just show all items and highlight directly
				for i := 0; i < visibleLines; i++ {
					if i < numItems {
						if i == selected {
							io.Display.WriteLineHighlighted(i, items[i])
						} else {
							io.Display.WriteLine(i, items[i])
						}
					} else {
						io.Display.WriteLine(i, "") // Clear unused lines
					}
				}
			} else {
				// Windowing logic for long menus
				start := selected - visibleLines/2
				if start < 0 {
					start = 0
				}
				if start > numItems-visibleLines {
					start = numItems - visibleLines
				}
				for i := 0; i < visibleLines; i++ {
					idx := start + i
					if idx < numItems {
						if i == (selected - start) {
							io.Display.WriteLineHighlighted(i, items[idx])
						} else {
							io.Display.WriteLine(i, items[idx])
						}
					} else {
						io.Display.WriteLine(i, "") // Clear unused lines
					}
				}
			}
			io.Display.Display()
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
		time.Sleep(20 * time.Millisecond)
	}
}
