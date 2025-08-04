package apps

import (
	"europi/buttons"
	"europi/controls"
	"europi/display"
	"image/color"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Pixels4 struct{}

func (Pixels4) Name() string { return "Pixels 4 Loop (v2)" }

// Rect is a simple rectangle struct.
type Rect struct{ X, Y, W, H int16 }

// rectsOverlapWithSpacing is a placeholder for collision detection.
func rectsOverlapWithSpacing(r1, r2 Rect) bool {
	// A simple overlap check with a 2-pixel spacing.
	return r1.X < r2.X+r2.W+2 && r1.X+r1.W+2 > r2.X &&
		r1.Y < r2.Y+r2.H+2 && r1.Y+r1.H+2 > r2.Y
}

// Helper function to draw rectangle outline
func drawRectangleOutline(ssd interface {
	SetPixel(x, y int16, c color.RGBA)
}, x, y, w, h int16) {
	// Top and bottom edges
	for i := int16(0); i < w; i++ {
		ssd.SetPixel(x+i, y, display.ColorWhite)
		ssd.SetPixel(x+i, y+h-1, display.ColorWhite)
	}
	// Left and right edges
	for i := int16(0); i < h; i++ {
		ssd.SetPixel(x, y+i, display.ColorWhite)
		ssd.SetPixel(x+w-1, y+i, display.ColorWhite)
	}
}

// plotLine is a placeholder for drawing a line.
func plotLine(ssd display.ISSD1306Device, x1, y1, x2, y2 int16) {
	// This is a basic implementation of Bresenham's line algorithm.
	dx := int16(math.Abs(float64(x2 - x1)))
	dy := int16(-math.Abs(float64(y2 - y1)))
	sx := int16(1)
	if x1 > x2 {
		sx = -1
	}
	sy := int16(1)
	if y1 > y2 {
		sy = -1
	}
	err := dx + dy
	for {
		ssd.SetPixel(x1, y1, display.ColorWhite)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
}

func (Pixels4) Run(hw *controls.Controls) {
	ssd, ok := hw.Display.GetSSD1306().(display.ISSD1306Device)
	if !ok {
		println("No SSD1306 device found, cannot run Pixels app")
		return
	}

	width, height := int16(128), int16(32)
	modes := []string{"Sine Wave", "Square Wave", "Random Lines", "Random Rectangles"}

	// state holds all shared data between the two goroutines.
	type state struct {
		mode          int
		knob2Value    int
		animationStep int16
		offset        int16
		frameDelay    time.Duration
		exit          bool
	}

	var mu sync.Mutex
	st := state{
		frameDelay:    100 * time.Millisecond,
		animationStep: 1,
	}

	// updateAnimParams calculates animation parameters based on the current state.
	// It must be called within a locked mutex.
	updateAnimParams := func() {
		k2 := st.knob2Value
		var delayMs int
		switch st.mode {
		case 0: // Sine Wave
			st.animationStep = int16(1 + k2/5)
			delayMs = 120 - k2/2
		case 1: // Square Wave
			st.animationStep = int16(1 + k2/10)
			delayMs = 100 - k2/3
		case 2: // Random Lines
			delayMs = 200 - k2
		case 3: // Random Rectangles
			minMs, maxMs := 10, 800
			delayMs = maxMs - (k2 * (maxMs - minMs) / 100)
		}

		// Ensure a minimum delay to prevent the ticker from panicking.
		if delayMs < 10 {
			delayMs = 10
		}
		st.frameDelay = time.Duration(delayMs) * time.Millisecond
	}
	// Set initial params
	updateAnimParams()

	// Goroutine for polling user inputs. It only writes to the state.
	go func() {
		btnMgr := buttons.New(hw.B1, hw.B2)
		lastKnob2Value := -1

		for {
			time.Sleep(20 * time.Millisecond) // Poll inputs at 50Hz

			mu.Lock()
			if st.exit {
				mu.Unlock()
				return // Exit goroutine
			}

			if btnMgr.BothHeld() {
				st.exit = true
				mu.Unlock()
				continue
			}

			k2 := hw.K2.Value()
			buttonEvent := btnMgr.Update()
			inputChanged := k2 != lastKnob2Value || buttonEvent != buttons.None

			if inputChanged {
				if k2 != lastKnob2Value {
					st.knob2Value = k2
					lastKnob2Value = k2
				}
				if buttonEvent == buttons.B1Press {
					st.mode = (st.mode - 1 + len(modes)) % len(modes)
					st.offset = 0 // Reset animation on mode change
				} else if buttonEvent == buttons.B2Press {
					st.mode = (st.mode + 1) % len(modes)
					st.offset = 0
				}
				updateAnimParams()
			}
			mu.Unlock()
		}
	}()

	// Main animation loop. It only reads from the state to draw frames.
	mu.Lock()
	currentDelay := st.frameDelay
	mu.Unlock()

	ticker := time.NewTicker(currentDelay)
	defer ticker.Stop()

	for {
		// 1. Wait for the next tick. The animation is now purely ticker-driven.
		<-ticker.C

		// 2. Lock the mutex and get a consistent snapshot of the state.
		mu.Lock()
		if st.exit {
			mu.Unlock()
			return // Exit the app
		}

		// Copy all necessary state into local variables for this frame.
		mode := st.mode
		offset := st.offset
		animationStep := st.animationStep
		knob2Value := st.knob2Value
		newDelay := st.frameDelay

		// Update the animation offset for the *next* frame.
		if mode == 0 || mode == 1 {
			st.offset = (offset + animationStep) % width
		}
		mu.Unlock()

		// 3. If the frame delay has changed, reset the ticker.
		if newDelay != currentDelay {
			ticker.Reset(newDelay)
			currentDelay = newDelay
		}

		// 4. Perform all drawing using the local state variables.
		ssd.ClearBuffer()
		switch mode {
		case 0: // Sine Wave
			for x := int16(0); x < width; x++ {
				xx := (x + offset) % width
				y := int16(16 + 12*math.Sin(float64(xx)*2*math.Pi/float64(width)))
				ssd.SetPixel(x, y, display.ColorWhite)
			}
		case 1: // Square Wave
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
			maxRects := 3 + int(knob2Value/20)
			for i := 0; i < maxRects*3 && len(placedRects) < maxRects; i++ {
				w := int16(8 + rand.Intn(12))
				h := int16(4 + rand.Intn(6))
				x := int16(rand.Intn(int(width - w)))
				y := int16(rand.Intn(int(height - h)))
				newRect := Rect{x, y, w, h}
				hasOverlap := false
				for _, existing := range placedRects {
					if rectsOverlapWithSpacing(newRect, existing) {
						hasOverlap = true
						break
					}
				}
				if !hasOverlap {
					if rand.Intn(3) == 0 {
						drawRectangleOutline(ssd, x, y, w, h)
					} else {
						ssd.FillRectangle(x, y, w, h, display.ColorWhite)
					}
					placedRects = append(placedRects, newRect)
				}
			}
		}
		ssd.Display()
	}
}