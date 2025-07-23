package pixel_animation

import (
	"europi/buttons"
	"europi/controls"
	"europi/display"
	"math"
	"math/rand"
	"time"
)

type Pixels struct{}

func (Pixels) Name() string { return "Pixels 1 Loop" }

func (Pixels) Run(hw *controls.Controls) {
	ssd, ok := hw.Display.GetSSD1306().(display.ISSD1306Device)
	if !ok {
		println("No SSD1306 device found or device does not support pixel operations, cannot run Pixels app")
		return
	}

	width, height := int16(128), int16(32)
	modes := []string{"Sine Wave", "Square Wave", "Random Lines", "Random Rectangles"}
	mode := 0

	var (
		animationStep int16 = 1
		frameDelay    time.Duration
		offset        int16
	)

	const loopDelay = 2 * time.Millisecond
	var lastFrameTime time.Time
	btnMgr := buttons.New(hw.B1, hw.B2)
	
	// Rapid loop and input handling
	var lastKnob2Value int = -1
	for {
		now := time.Now()

		// -- Fast Input Check --
		switch btnMgr.Update() {
		case buttons.B1Press:
			mode--
			if mode < 0 {
				mode = len(modes) - 1
			}
			offset = 0
		case buttons.B2Press:
			mode++
			if mode >= len(modes) {
				mode = 0
			}
			offset = 0
		}
	
		if btnMgr.BothHeld() {
			println("Exiting due to long press")
			return
		}

		knob2Value := hw.K2.Value()
		if knob2Value != lastKnob2Value {
			switch mode {
			case 0:
				animationStep = int16(1 + knob2Value/5)
				frameDelay = time.Duration(120-knob2Value/2) * time.Millisecond
			case 1:
				animationStep = int16(1 + knob2Value/10)
				frameDelay = time.Duration(100-knob2Value/3) * time.Millisecond
			case 2:
				frameDelay = time.Duration(200-knob2Value) * time.Millisecond
			case 3:
				minMs, maxMs := 10, 800 // Fastest: 10ms, Slowest: 800ms
				k2 := int(knob2Value)
				ms := maxMs - (k2 * (maxMs - minMs) / 100)
				if ms < minMs {
					ms = minMs
				}
				if ms > maxMs {
					ms = maxMs
				}
				frameDelay = time.Duration(ms) * time.Millisecond
			}
			lastKnob2Value = knob2Value
		}

		// -- Timed Frame Update --
		if now.Sub(lastFrameTime) >= frameDelay {
			lastFrameTime = now

			ssd.ClearBuffer()

			switch mode {
			case 0: // Sine
				for x := int16(0); x < width; x++ {
					xx := (x + offset) % width
					y := int16(16 + 12*math.Sin(float64(xx)*2*math.Pi/float64(width)))
					if y >= 0 && y < height {
						ssd.SetPixel(x, y, display.ColorWhite)
					}
				}
			case 1: // Square
				highY, lowY := int16(8), int16(24)
				period := int16(32)
				half := period / 2
				for x := int16(0); x < width; x++ {
					xx := (x + offset) % period
					y := lowY
					if xx < half {
						y = highY
					}
					ssd.SetPixel(x, y, display.ColorWhite)
					if xx == 0 || xx == half {
						for y2 := highY; y2 <= lowY; y2++ {
							ssd.SetPixel(x, y2, display.ColorWhite)
						}
					}
				}
			case 2: // Random Lines
				numLines := 5 + int(knob2Value/25)
				for i := 0; i < numLines; i++ {
					x1 := int16(rand.Intn(int(width)))
					y1 := int16(rand.Intn(int(height)))
					x2 := int16(rand.Intn(int(width)))
					y2 := int16(rand.Intn(int(height)))
					plotLine(ssd, x1, y1, x2, y2)
				}
			case 3: // Random Rectangles
				var placedRects []Rect
				maxRects := 3 + int(knob2Value/50)
				attempts := 0
				maxAttempts := 50
				for len(placedRects) < maxRects && attempts < maxAttempts {
					x := int16(rand.Intn(int(width - 20)))
					y := int16(rand.Intn(int(height - 10)))
					w := int16(8 + rand.Intn(12))
					h := int16(4 + rand.Intn(6))
					// Ensure rectangles stay within bounds
					if x+w > width {
						w = width - x
					}
					if y+h > height {
						h = height - y
					}
					newRect := Rect{x, y, w, h}
					// Check for overlap with 1-pixel spacing
					hasOverlap := false
					for _, existing := range placedRects {
						if rectsOverlapWithSpacing(newRect, existing) {
							hasOverlap = true
							break
						}
					}
					if !hasOverlap {
						// Randomly choose filled or outline
						if rand.Intn(3) == 0 {
							drawRectangleOutline(ssd, x, y, w, h)
						} else {
							ssd.FillRectangle(x, y, w, h, display.ColorWhite)
						}
						placedRects = append(placedRects, newRect)
					}
					attempts++
				}
			}

			ssd.Display()

			if mode == 0 || mode == 1 {
				offset = (offset + animationStep) % width
			}
		}

		time.Sleep(loopDelay)
	}
}
