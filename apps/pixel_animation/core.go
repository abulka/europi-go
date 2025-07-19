package pixel_animation

import (
	"europi/buttons"
	"europi/controls"
	"europi/display"
	"europi/scheduler"
	"image/color"
	"math"
	"math/rand"
	"time"
)

// Type assert to the SSD1306 device interface
type SSD1306Device interface {
	SetPixel(x, y int16, c color.RGBA)
	Display() error
	ClearDisplay()
	ClearBuffer()
	FillRectangle(x, y, width, height int16, c color.RGBA) error
}

type AnimationState struct {
	hw             *controls.Controls
	ssd            SSD1306Device
	scheduler      *scheduler.Scheduler
	width, height  int16
	mode           int
	modes          []string
	offset         int16
	animationStep  int16
	frameDelay     time.Duration
	lastKnob2Value int
	btnMgr         *buttons.ButtonManager
	running        bool
}

func (state *AnimationState) startCurrentAnimation() {
	state.updateFrameDelay()

	var cb func()
	switch state.mode {
	case 0:
		cb = func() { sineWaveAnimation(state) }
	case 1:
		cb = func() { squareWaveAnimation(state) }
	case 2:
		cb = func() { randomLinesAnimation(state) }
	case 3:
		cb = func() { randomRectanglesAnimation(state) }
	}

	// Remove old animation task
	state.scheduler.RemoveTask("animation", false)

	// Run the new animation immediately, it will auto reschedule itself
	cb()
}

func (state *AnimationState) updateFrameDelay() {
	knob2Value := state.hw.K2.Value()

	switch state.mode {
	case 0: // Sine Wave
		state.animationStep = int16(1 + knob2Value/5)
		state.frameDelay = time.Duration(120-knob2Value/2) * time.Millisecond
	case 1: // Square Wave
		state.animationStep = int16(1 + knob2Value/10)
		state.frameDelay = time.Duration(100-knob2Value/3) * time.Millisecond
	case 2: // Random Lines
		state.frameDelay = time.Duration(200-knob2Value) * time.Millisecond
	case 3: // Random Rectangles
		minMs, maxMs := 10, 800
		k2 := int(knob2Value)
		ms := maxMs - (k2 * (maxMs - minMs) / 100)
		if ms < minMs {
			ms = minMs
		}
		if ms > maxMs {
			ms = maxMs
		}
		state.frameDelay = time.Duration(ms) * time.Millisecond
	}
}

func sineWaveAnimation(state *AnimationState) {
	if !state.running {
		return
	}

	state.ssd.ClearBuffer()

	for x := int16(0); x < state.width; x++ {
		xx := (x + state.offset) % state.width
		y := int16(16 + 12*math.Sin(float64(xx)*2*math.Pi/float64(state.width)))
		if y >= 0 && y < state.height {
			state.ssd.SetPixel(x, y, display.ColorWhite)
		}
	}

	state.ssd.Display()
	state.offset = (state.offset + state.animationStep) % state.width

	// Reschedule next frame
	state.scheduler.AddTaskWithName(func() { sineWaveAnimation(state) }, state.frameDelay, "animation")
}

func squareWaveAnimation(state *AnimationState) {
	if !state.running {
		return
	}

	state.ssd.ClearBuffer()

	highY, lowY := int16(8), int16(24)
	period := int16(32)
	half := period / 2

	for x := int16(0); x < state.width; x++ {
		xx := (x + state.offset) % period
		y := lowY
		if xx < half {
			y = highY
		}
		state.ssd.SetPixel(x, y, display.ColorWhite)
		if xx == 0 || xx == half {
			for y2 := highY; y2 <= lowY; y2++ {
				state.ssd.SetPixel(x, y2, display.ColorWhite)
			}
		}
	}

	state.ssd.Display()
	state.offset = (state.offset + state.animationStep) % state.width

	// Reschedule next frame
	state.scheduler.AddTaskWithName(func() { squareWaveAnimation(state) }, state.frameDelay, "animation")
}

func randomLinesAnimation(state *AnimationState) {
	if !state.running {
		return
	}

	state.ssd.ClearBuffer()

	numLines := 5 + int(state.hw.K2.Value()/25)
	for i := 0; i < numLines; i++ {
		x1 := int16(rand.Intn(int(state.width)))
		y1 := int16(rand.Intn(int(state.height)))
		x2 := int16(rand.Intn(int(state.width)))
		y2 := int16(rand.Intn(int(state.height)))
		plotLine(state.ssd, x1, y1, x2, y2)
	}

	state.ssd.Display()

	// Reschedule next frame
	state.scheduler.AddTaskWithName(func() { randomLinesAnimation(state) }, state.frameDelay, "animation")
}

func randomRectanglesAnimation(state *AnimationState) {
	if !state.running {
		return
	}

	state.ssd.ClearBuffer()

	var placedRects []Rect
	maxRects := 3 + int(state.hw.K2.Value()/50)
	attempts := 0
	maxAttempts := 50

	for len(placedRects) < maxRects && attempts < maxAttempts {
		x := int16(rand.Intn(int(state.width - 20)))
		y := int16(rand.Intn(int(state.height - 10)))
		w := int16(8 + rand.Intn(12))
		h := int16(4 + rand.Intn(6))

		// Ensure rectangles stay within bounds
		if x+w > state.width {
			w = state.width - x
		}
		if y+h > state.height {
			h = state.height - y
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
				drawRectangleOutline(state.ssd, x, y, w, h)
			} else {
				state.ssd.FillRectangle(x, y, w, h, display.ColorWhite)
			}
			placedRects = append(placedRects, newRect)
		}
		attempts++
	}

	state.ssd.Display()

	// Reschedule next frame
	state.scheduler.AddTaskWithName(func() { randomRectanglesAnimation(state) }, state.frameDelay, "animation")
}

// Bresenham's line algorithm for RGBA color display
func plotLine(ssd interface {
	SetPixel(x, y int16, c color.RGBA)
}, x0, y0, x1, y1 int16) {
	dx := abs16(x1 - x0)
	sx := int16(1)
	if x0 > x1 {
		sx = -1
	}
	dy := -abs16(y1 - y0)
	sy := int16(1)
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		ssd.SetPixel(x0, y0, display.ColorWhite)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
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

// Helper function to check rectangle overlap with 1-pixel spacing
func rectsOverlapWithSpacing(a, b Rect) bool {
	// Expand each rectangle by 1 pixel on all sides for spacing check
	return a.X-1 < b.X+b.W && a.X+a.W+1 > b.X &&
		a.Y-1 < b.Y+b.H && a.Y+a.H+1 > b.Y
}

func abs16(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}

type Rect struct {
	X, Y, W, H int16
}
