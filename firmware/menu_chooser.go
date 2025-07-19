package firmware

import (
	"europi/controls"
	"time"
)

// MenuChooser displays a scrollable menu of registered apps, allows selection with K2, launch with B2
func MenuChooser(hw *controls.Controls, visibleLines int) int {
	numApps := len(appRegistry)
	if numApps == 0 {
		return -1
	}
	names := make([]string, numApps)
	for i, app := range appRegistry {
		names[i] = app.Name()
	}
	return ScrollingMenu(names, hw, visibleLines)
}

// ScrollingMenu displays a scrollable menu of items, allows selection with K2, launch with B2
// Returns the selected index, or -1 if exited
func ScrollingMenu(items []string, hw *controls.Controls, visibleLines int) int {
	numItems := len(items)
	if numItems == 0 {
		return -1
	}
	// Insert menu header at the top
	menuItems := make([]string, numItems+1)
	menuItems[0] = "--- MENU ---"
	copy(menuItems[1:], items)
	totalItems := numItems + 1
	selected := 1 // Start at first selectable item
	selectedLast := -1
	lastK2 := -1
	for {
		k2 := hw.K2.Value()
		updateDisplay := false
		if k2 != lastK2 {
			bins := totalItems - 1
			selected = 1 + ((k2 * bins) / 101)
			if selected < 1 {
				selected = 1
			}
			if selected >= totalItems {
				selected = totalItems - 1
			}
			lastK2 = k2
			if selected != selectedLast {
				selectedLast = selected
				updateDisplay = true
			}
		}
		if updateDisplay {
			hw.Display.ClearBuffer()
			if totalItems <= visibleLines {
				for i := 0; i < visibleLines; i++ {
					if i < totalItems {
						if i == selected && i != 0 {
							hw.Display.WriteLineHighlighted(i, menuItems[i])
						} else {
							hw.Display.WriteLine(i, menuItems[i])
						}
					} else {
						hw.Display.WriteLine(i, "")
					}
				}
			} else {
				start := selected - visibleLines/2
				if start < 0 {
					start = 0
				}
				if start > totalItems-visibleLines {
					start = totalItems - visibleLines
				}
				for i := 0; i < visibleLines; i++ {
					idx := start + i
					if idx < totalItems {
						if idx == selected && idx != 0 {
							hw.Display.WriteLineHighlighted(i, menuItems[idx])
						} else {
							hw.Display.WriteLine(i, menuItems[idx])
						}
					} else {
						hw.Display.WriteLine(i, "")
					}
				}
			}
			hw.Display.Display()
		}

		if hw.B2.Pressed() && !hw.B1.Pressed() {
			for hw.B2.Pressed() {
				time.Sleep(10 * time.Millisecond)
			}
			// Return selected-1 so 0 is first app, etc. Never return 0 (header)
			return selected - 1
		}
		if ShouldExit(hw) {
			return -1
		}
		time.Sleep(2 * time.Millisecond)
	}
}
