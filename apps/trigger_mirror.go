package apps

import (
	"europi/buttons"
	"europi/controls"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// TriggerMirror: sets CV1 high on DIN rise, low on DIN fall.
type TriggerMirror struct{}

func (TriggerMirror) Name() string { return "Trigger Mirror" }

type TMState struct {
	hw          *controls.Controls
	running     bool
	btnMgr      *buttons.ButtonManager
	gateRunning bool
	gateIsHigh  bool
	uniqueId    int
	edgeEvents  chan bool // true for rise, false for fall
}

func (TriggerMirror) Run(hw *controls.Controls) {
	state := &TMState{
		hw:          hw,
		running:     true,
		gateRunning: true,
		btnMgr:      buttons.New(hw.B1, hw.B2),
		gateIsHigh:  false,
		uniqueId:    rand.Int(),
		edgeEvents:  make(chan bool, 8),
	}

	// This helper function sends an event from the ISR to the channel.
	// The select/default pattern guarantees the ISR never blocks, even if the channel were full.
	sendEvent := func(isRise bool) {
		select {
		case state.edgeEvents <- isRise:
			// Event sent
		default:
			// Channel full, event dropped (safeguard)
		}
	}

	hw.DIN.SetEdgeHandlers(
		func() { // Rise handler
			if state.gateRunning {
				sendEvent(true)
			}
		},
		func() { // Fall handler
			if state.gateRunning {
				sendEvent(false)
			}
		},
	)

	// Launch drawscreen in a separate goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for state.running {
			state.drawScreen()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// *** REVISED MAIN LOOP ***
	// This loop continuously runs without a long sleep, making it highly responsive.
	for state.running {

		// Step 1: Perform a NON-BLOCKING check for DIN events.
		// The 'default' case means the loop doesn't wait here if the channel is empty.
		select {
		case isRise := <-state.edgeEvents:
			if isRise {
				state.hw.CV1.On()
				state.gateIsHigh = true
			} else {
				state.hw.CV1.Off()
				state.gateIsHigh = false
			}
		default:
			// No event waiting, so we immediately continue to the next step.
		}

		// Step 2: Check the buttons on EVERY loop iteration.
		// This is no longer dependent on a ticker and will be checked constantly.
		switch state.btnMgr.Update() {
		case buttons.B1Press:
			state.gateRunning = !state.gateRunning
		}

		// Check for the exit condition. Note: BothHeld() means hold B1 and B2 down
		// at the same time, not a double press.
		if state.btnMgr.BothHeld() {
			state.gateRunning = false
			state.running = false
		}
	}

	// Cleanup
	hw.DIN.UnsetInterrupt()
	wg.Wait()
}

// drawScreen function remains the same.
func (s *TMState) drawScreen() {
	s.hw.Display.ClearBuffer()
	s.hw.Display.WriteLine(0, "Trigger Mirror 2")
	s.hw.Display.WriteLine(1, "Running: "+strconv.FormatBool(s.gateRunning))
	if s.gateIsHigh {
		s.hw.Display.WriteLine(2, "CV1 is ON")
	} else {
		s.hw.Display.WriteLine(2, "CV1 is OFF")
	}
	s.hw.Display.WriteLine(3, "ID: "+strconv.Itoa(s.uniqueId))
	s.hw.Display.Display()
}
