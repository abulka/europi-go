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

	selected := 0
	lastK2 := io.K2.Value()

	// When using tinyfont mode where we write text to the screen using the
	// tinyfont library, we need to ClearDisplay() before writing different text
	// (like changing the same line from highlighted to non highlighted), as
	// this is the only way to truly erase old text content. Weirdly, using
	// display.FillRectangle(x, y, w, h, c) to erase existing lines of text does
	// not work. This horrible, inefficient redrawing of the screen each time is mitigated
	// by the buffered decorator, which only redraws the screen if the content has changed.
	for {
		k2 := io.K2.Value()
		updateDisplay := false
		if k2 != lastK2 {
			selected = (k2 * numApps) / 100
			if selected < 0 {
				selected = 0
			}
			if selected >= numApps {
				selected = numApps - 1
			}
			lastK2 = k2
			updateDisplay = true
		}

		// Only update the display if K2 changed, or on the first loop
		if updateDisplay {
			io.Display.ClearDisplay()
			if numApps <= visibleLines {
				// No windowing, just show all items and highlight directly
				for i := 0; i < visibleLines; i++ {
					if i < numApps {
						if i == selected {
							io.Display.WriteLineHighlighted(i, appRegistry[i].Name())
						} else {
							io.Display.WriteLine(i, appRegistry[i].Name())
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
				if start > numApps-visibleLines {
					start = numApps - visibleLines
				}
				for i := 0; i < visibleLines; i++ {
					idx := start + i
					if idx < numApps {
						if i == (selected - start) {
							io.Display.WriteLineHighlighted(i, appRegistry[idx].Name())
						} else {
							io.Display.WriteLine(i, appRegistry[idx].Name())
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
		time.Sleep(30 * time.Millisecond)
	}
}
