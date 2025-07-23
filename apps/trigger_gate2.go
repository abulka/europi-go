package apps

import (
	"europi/buttons"
	"europi/controls"
	"fmt"
	"runtime"
	"sync"
	"time"
)

/*
Trigger Gate Delay
author: Andy Bulka (tcab) (github.com/abulka)
date: 2025-07-20
labels: trigger, gate, delay

Generates a gate on cv1 in response to a trigger on din.
Control the outgoing pulse width with k1. Control the delay between the trigger
and the gate starting with k2. Handy for converting short triggers (e.g. 1ms)
into longer gates (e.g. 10ms) as some eurorack modules don't like short
triggers.

TriggerGateDelay2 is a version of TriggerGateDelay that uses a time-based approach
instead of a scheduler. It tracks the next gate on and off times using time.Time fields.
This allows for precise timing without the overhead of a scheduler, making it suitable
for applications where timing accuracy is critical.
*/

type TriggerGateDelay2 struct{}

func (TriggerGateDelay2) Name() string { return "Trigger Gate 2" }

// TGDState2 holds the application's state. The scheduler has been replaced
// with time.Time fields to track when future events should occur.
type TGDState2 struct {
	hw *controls.Controls

	// --- Timing State ---
	// These fields replace the scheduler. A zero time.Time means no event is scheduled.
	nextGateOnTime  time.Time
	nextGateOffTime time.Time

	// --- Digital Input State ---
	dinPulseWidth time.Duration
	dinPeriod     time.Duration
	dinHz         float64
	lastTrigger   time.Time
	edgeEvents    chan bool // Channel to receive edge events from the ISR

	// --- Gate Output State ---
	gateIsHigh         bool // True if the gate is currently high
	gateRunning        bool // True if the gate logic is enabled
	gateDelay          time.Duration
	gatePulseWidth     time.Duration
	afterOffSettlingMs time.Duration // Time to wait after forcing a gate off before a new one can start

	// --- UI & Control State ---
	btnMgr         *buttons.ButtonManager
	running        bool
	updateUI       bool
	lastK1, lastK2 int

	// --- Gate Timing Diagnostics ---
	prevGateOnTime  time.Time
	prevGateOffTime time.Time
	expectedGateMs  int
}

// max returns the larger of two time.Duration values.
func max(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// Dummy implementation for calcHz. Replace with your actual implementation.
func calcHz2(s *TGDState2) {
	if s.dinPeriod > 0 {
		s.dinHz = 1.0 / s.dinPeriod.Seconds()
	} else {
		s.dinHz = 0
	}
}

// Run is the main entry point for the application.
func (TriggerGateDelay2) Run(hw *controls.Controls) {

	// Initialize the application state.
	state := &TGDState2{
		hw:                 hw,
		running:            true,
		btnMgr:             buttons.New(hw.B1, hw.B2),
		edgeEvents:         make(chan bool, 8),
		afterOffSettlingMs: 1 * time.Millisecond, // 1ms settling time.
	}

	// WaitGroup to manage background goroutines for clean shutdown.
	var wg sync.WaitGroup

	// Defer the cleanup logic to run when the 'Run' function exits.
	defer func() {
		println("Cleaning up...")
		// Disable the hardware interrupt to prevent it from firing after the app has exited.
		hw.DIN.UnsetInterrupt()
		// Wait for all background goroutines managed by the WaitGroup to finish.
		wg.Wait()
		println("Cleanup complete.")
	}()

	// Set initial values.
	state.gateRunning = true
	state.gatePulseWidth = 10 * time.Millisecond
	state.gateDelay = 0
	state.updateUI = true // Force initial screen draw

	// This helper function sends an event from the ISR to the channel.
	// The select/default pattern guarantees the ISR never blocks.
	sendEvent := func(isRise bool) {
		select {
		case state.edgeEvents <- isRise:
			// Event sent successfully.
		default:
			// Channel was full, so the event is dropped.
			// This is a safeguard against the main loop getting stuck.
		}
	}

	// Configure the digital input to send events to our channel.
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

	// --- Background Goroutine for Low-Priority Tasks ---
	// This goroutine handles UI updates and state saving, preventing them
	wg.Add(1)
	go func() {
		defer wg.Done()
		lastDisplay := time.Now()
		lastSave := time.Now()

		for state.running {
			now := time.Now()

			if now.Sub(lastDisplay) > 150*time.Millisecond {
				state.drawScreen()
				lastDisplay = now
			}
			if now.Sub(lastSave) > 5*time.Second {
				state.saveState()
				lastSave = now
			}

			runtime.Gosched()
		}
	}()


	// --- Main Application Loop (High-Priority Tasks Only) ---
	for state.running {
		now := time.Now()

		// --- 1. Process Time-Critical Events ---
		// Check if it's time to turn the gate on.
		if !state.nextGateOnTime.IsZero() && now.After(state.nextGateOnTime) {
			state.gateOn()
			state.nextGateOnTime = time.Time{} // Clear the event so it doesn't re-fire.
		}

		// Check if it's time to turn the gate off.
		if !state.nextGateOffTime.IsZero() && now.After(state.nextGateOffTime) {
			state.gateOff()
			state.nextGateOffTime = time.Time{} // Clear the event.
		}

		// --- 2. Process Asynchronous Inputs (Digital In) ---
		// Perform a non-blocking check for DIN events.
		select {
		case isRise := <-state.edgeEvents:
			if isRise {
				triggerTime := time.Now()
				if !state.lastTrigger.IsZero() {
					state.dinPeriod = triggerTime.Sub(state.lastTrigger)
					calcHz2(state)
					state.updateUI = true
				}
				state.lastTrigger = triggerTime

				delay := state.gateDelay
				// If a trigger arrives while the gate is already high (a re-trigger)...
				if state.gateIsHigh {
					state.gateOff()                     // ...turn the current gate off immediately.
					state.nextGateOffTime = time.Time{} // ...and cancel its scheduled turn-off time.
					// Use a small settling delay to ensure the output signal falls cleanly.
					delay = max(state.gateDelay, state.afterOffSettlingMs)
				}

				// Schedule the new gate to turn on after the calculated delay.
				if state.gateRunning {
					state.nextGateOnTime = triggerTime.Add(delay)
				}

			} else {
				// Falling edge: calculate the input pulse width.
				state.dinPulseWidth = time.Since(state.lastTrigger)
			}
		default:
			// No event waiting, so we immediately continue.
		}

		// --- 3. Process Polled Inputs (Buttons & Knobs) ---
		switch state.btnMgr.Update() {
		case buttons.B1Press:
			state.gateRunning = !state.gateRunning
			state.updateUI = true
		}

		if state.btnMgr.BothHeld() {
			println("Exiting due to long press")
			state.running = false
			// return // Exit the Run function
			continue // Skip to the top to exit loop
		}

		k1 := state.hw.K1.Value()
		if k1 != state.lastK1 {
			state.gatePulseWidth = time.Duration(k1) * time.Millisecond
			state.lastK1 = k1
			state.updateUI = true
		}

		k2 := state.hw.K2.Value()
		if k2 != state.lastK2 {
			state.gateDelay = time.Duration(k2) * time.Millisecond
			state.lastK2 = k2
			state.updateUI = true
		}

		// Yield to other goroutines. This is good practice in a tight loop.
		runtime.Gosched()
	}
}

// gateOn turns the gate output high and schedules it to turn off.
func (s *TGDState2) gateOn() {
	now := time.Now()

	// --- Sanity Checks ---
	if !s.prevGateOffTime.IsZero() && now.Sub(s.prevGateOffTime) < s.gatePulseWidth {
		fmt.Printf("[WARN] Gate retriggered %dms after gate off\n", now.Sub(s.prevGateOffTime).Milliseconds())
	}
	if !s.lastTrigger.IsZero() && now.Sub(s.lastTrigger) > s.gateDelay+(5*time.Millisecond) {
		fmt.Printf("[ERROR] Gate triggered too late after last trigger: by %dms\n", now.Sub(s.lastTrigger).Milliseconds())
	}

	s.prevGateOnTime = now
	s.gateIsHigh = true
	s.hw.CV1.On()

	// For diagnostics
	s.expectedGateMs = int(s.gatePulseWidth.Milliseconds())

	// Schedule the gate to turn off.
	s.nextGateOffTime = now.Add(s.gatePulseWidth)
}

// gateOff turns the gate output low.
func (s *TGDState2) gateOff() {
	now := time.Now()
	duration := now.Sub(s.prevGateOnTime).Milliseconds()
	s.prevGateOffTime = now

	// --- Sanity Checks ---
	if duration > int64(s.expectedGateMs)+2 {
		fmt.Printf("[ERROR] Gate lasted too long: %dms > expected %dms\n", duration, s.expectedGateMs)
	}
	if duration < int64(s.expectedGateMs)-2 {
		fmt.Printf("[ERROR] Gate too short: %dms < expected %dms\n", duration, s.expectedGateMs)
	}

	s.hw.CV1.Off()
	s.gateIsHigh = false
}

// drawScreen updates the OLED display if the state has changed.
func (s *TGDState2) drawScreen() {
	if s.updateUI {
		// println("Drawing screen...")
		s.hw.Display.ClearBuffer()
		isRunning := ""
		if s.gateRunning {
			isRunning = "."
		}
		s.hw.Display.WriteLine(0, fmt.Sprintf("DIN Pw %dms %s", s.dinPulseWidth.Milliseconds(), isRunning))
		s.hw.Display.WriteLine(1, fmt.Sprintf("  %.1fHz %dms", s.dinHz, s.dinPeriod.Milliseconds()))
		s.hw.Display.WriteLine(2, fmt.Sprintf("GATE %dms Dly %dms", s.gatePulseWidth.Milliseconds(), s.gateDelay.Milliseconds()))
		s.hw.Display.Display()
		s.updateUI = false
	}
	// Note: Rescheduling is now handled in the main loop.
}

// saveState is a stub for saving persistent state.
func (s *TGDState2) saveState() {
	// fmt.Println("Saving state...")
	// Note: Rescheduling is now handled in the main loop.
}
