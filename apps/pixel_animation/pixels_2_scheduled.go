package pixel_animation

import (
	"europi/buttons"
	"europi/controls"
	"europi/display"
	"europi/scheduler"
	"sync"
	"time"
)

type Pixels2 struct{}

func (Pixels2) Name() string { return "Pixels 2 Sched" }

func (Pixels2) Run(hw *controls.Controls) {
	ssd, ok := hw.Display.GetSSD1306().(display.ISSD1306Device)
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
		offset:         0,
		animationStep:  1,
		frameDelay:     100,
		lastKnob2Value: -1,
		btnMgr:         buttons.New(hw.B1, hw.B2),
		running:        true,
	}

	// Start the input handler
	state.scheduler.AddTask(func() { inputHandler(state) }, 2*time.Millisecond, "input_handler")

	// Start the first animation
	state.startCurrentAnimation()

	// Run the scheduler with WaitGroup for clean shutdown
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		state.scheduler.Run()
	}()
	defer func() {
		state.scheduler.Stop()
		wg.Wait() // Wait for scheduler goroutine to exit
		println("Scheduler stopped, exiting Pixels 2 app")
	}()

	// Main loop just waits for exit condition
	for state.running {
		time.Sleep(10 * time.Millisecond)
	}

}

func inputHandler(state *AnimationState) {
	if !state.running {
		return
	}

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

	k2Val := state.hw.K2.Value()
	if k2Val != state.lastKnob2Value {
		state.lastKnob2Value = k2Val
		state.updateFrameDelay()
	}

	// Reschedule input handler
	state.scheduler.AddTask(func() { inputHandler(state) }, 2*time.Millisecond, "input_handler")
}
