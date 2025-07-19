package pixel_animation

import (
	"europi/buttons"
	"europi/controls"
	"europi/scheduler"
	"time"
)

/*
This Pixels app uses concurrent goroutines for responsive input handling and
animation scheduling.

Main goroutine: The main app loop initializes the system, launches the
scheduler, and handles input events in a fast loop for responsiveness. It
communicates with the scheduler to change animations based on button presses and
knob changes.

Scheduler goroutine: Manages the timing and running of the animation tasks.
*/

type Pixels3 struct{}

func (Pixels3) Name() string { return "Pixels 3 FastIO" }

func (Pixels3) Run(hw *controls.Controls) {
	ssd, ok := hw.Display.GetSSD1306().(SSD1306Device)
	if !ok {
		println("No SSD1306 device found or device does not support pixel operations, cannot run Pixels app")
		return
	}

	state := &AnimationState{
		hw:             hw,
		ssd:            ssd,
		scheduler:      scheduler.New(),
		width:          128,
		height:         32,
		modes:          []string{"Sine Wave", "Square Wave", "Random Lines", "Random Rectangles"},
		mode:           0,
		animationStep:  1,
		frameDelay:     100 * time.Millisecond,
		lastKnob2Value: -1,
		btnMgr:         buttons.New(hw.B1, hw.B2),
		running:        true,
	}

	// Start the animation and scheduler
	state.startCurrentAnimation()
	go state.scheduler.Run()
	defer func() {
		state.running = false
		state.scheduler.Stop()
	}()

	// Input processing loop
	for state.running {
		switch state.btnMgr.Update() {
		case buttons.B1Press:
			state.mode--
			if state.mode < 0 {
				state.mode = len(state.modes) - 1
			}
			state.startCurrentAnimation()

		case buttons.B2Press:
			state.mode++
			if state.mode >= len(state.modes) {
				state.mode = 0
			}
			state.startCurrentAnimation()
		}

		if state.btnMgr.BothHeld() {
			println("Exiting due to long press")
			state.running = false
			return
		}

		k2Val := hw.K2.Value()
		if k2Val != state.lastKnob2Value {
			state.lastKnob2Value = k2Val
			state.updateFrameDelay()
		}

		time.Sleep(5 * time.Millisecond)
	}
}
