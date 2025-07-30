package apps

import (
	"europi/buttons"
	"europi/controls"
	scheduler "europi/schedulerc"
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
*/

type TriggerGateDelay struct{}

func (TriggerGateDelay) Name() string { return "Trigger to Gate" }

type TGDState struct {
	prevGateOnTime  time.Time
	prevGateOffTime time.Time
	expectedGateMs  int // for tracking expected duration

	hw                 *controls.Controls
	scheduler          *scheduler.ChannelScheduler
	running            bool
	btnMgr             *buttons.ButtonManager
	dinPulseWidth      time.Duration
	dinPeriod          time.Duration
	dinHz              float64 // Period in Hz
	lastTrigger        time.Time
	gateIsHigh         bool // True if the gate is currently high
	updateUI           bool
	lastK1, lastK2     int
	edgeEvents         chan bool   // true for rise, false for fall
	taskChan           chan func() // channel for tasks to run
	taskChanLow        chan func() // channel for low-priority tasks to run
	gateRunning        bool
	gateDelay          time.Duration
	gatePulseWidth     time.Duration
	knob1Range         []int
	knob2Range         []int
	afterOffSettling   time.Duration // time to wait after turning off gate or clock output, or it doesn't happen cleanly
	debug              bool          // Enable debug output
}

/*
Uses ChannelScheduler to run tasks in a separate goroutine (the normal mutex based scheduler crashed).
1. goroutine: Task Scheduler
2. goroutine: task channel executor for high-priority tasks
3. goroutine: task channel executor for low-priority tasks
4. main: main loop handles interrupts (creating tasks) and the user hardware controls (knobs and buttons).
*/
func (TriggerGateDelay) Run(hw *controls.Controls) {

	state := &TGDState{
		hw:          hw,
		running:     true,
		btnMgr:      buttons.New(hw.B1, hw.B2),
		knob1Range:  buildRange(),
		knob2Range:  append([]int{0}, buildRange()...),
		edgeEvents:  make(chan bool, 8),
		taskChan:    make(chan func(), 16),
		taskChanLow: make(chan func(), 16),
		debug:       false,
	}
	// Create channel-only scheduler
	state.scheduler = scheduler.NewChannelScheduler(state.taskChan, state.taskChanLow)

	// WaitGroup to manage background goroutines for clean shutdown.
	var wg sync.WaitGroup

	// Defer the cleanup logic to run when the 'Run' function exits.
	defer func() {
		println("Cleaning up...")
		// Disable the hardware interrupt to prevent it from firing after the app has exited.
		hw.DIN.UnsetInterrupt()

		// Signal the goroutine to stop by setting 'running' to false
		state.running = false // Set this *before* closing the channel

		// Close the task channel to unblock the goroutine
		close(state.taskChan)    // This is the key!
		close(state.taskChanLow) // Close the low-priority task channel too

		// Wait for all background goroutines managed by the WaitGroup to finish.
		println("Waiting for background tasks to finish...")
		wg.Wait()
		println("Cleanup complete.")
	}()

	// Initial values
	state.gateRunning = true
	state.gatePulseWidth = 10 * time.Millisecond
	state.gateDelay = 0

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

	wg.Add(1)
	go state.scheduler.Run(&wg) // Start the scheduler in a separate goroutine
	defer state.scheduler.Stop()
	// time.Sleep(1 * time.Second) // Allow time for the scheduler to start for TinyGo

	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range state.taskChan {
			// print("  Exec task ")
			task() // executes safely without holding scheduler lock
			if !state.running {
				println("Task channel closed, exiting background goroutine.")
				return
			}
			runtime.Gosched() // Yield to allow other goroutines to run
		}
	}()
	// time.Sleep(1 * time.Second) // Allow time for the go routine to start for TinyGo

	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range state.taskChanLow {
			// print("  Exec low-priority task ")
			task() // executes safely without holding scheduler lock
			if !state.running {
				println("Task channel (low priority) closed, exiting background goroutine.")
				return
			}
			runtime.Gosched() // Yield to allow other goroutines to run
		}
	}()
	// time.Sleep(1 * time.Second) // Allow time for the go routine to start for TinyGo

	state.scheduler.AddTask(state.drawScreen, 0*time.Millisecond, "low_drawScreen")
	state.scheduler.AddTask(state.saveState, 5*time.Second, "saveState")

	// Main Loop, Blocks on channels, timer used to check for user controls
	var timer *time.Timer       // pointer to a time.Timer object
	var timerC <-chan time.Time // receive-only channel of type time.Time
	for state.running {

		nextWake := 30 * time.Millisecond // 10ms is too fast and crashes

		if timer == nil {
			timer = time.NewTimer(nextWake) // timer has a channel field called C
			timerC = timer.C                // channel receives a value (the current time) when the timer expires.
		} else {
			timer.Stop()
			timer.Reset(nextWake) // Update the timer to the new duration wanted
		}

		select {
		case isRise := <-state.edgeEvents:
			state.handleDinEvent(isRise)
		case <-timerC:

			// Handle button presses
			switch state.btnMgr.Update() {
			case buttons.B1Press:
				state.gateRunning = !state.gateRunning
				state.updateUI = true
			}

			// Handle knob changes
			state.handleKnobChanges()

			if state.btnMgr.BothHeld() {
				println("Exiting due to long press")
				state.running = false
				continue
			}

		}

		time.Sleep(1 * time.Millisecond) // runtime.Gosched() doesn't work
	} // End of main loop
}

func (s *TGDState) handleKnobChanges() {
	// Handle K1 - gate pulse width
	k1 := s.hw.K1.Value() // Assuming this returns 0-1023 or similar
	if k1 != s.lastK1 {
		s.gatePulseWidth = time.Duration(k1) * time.Millisecond
		s.lastK1 = k1
		s.updateUI = true
	}

	// Handle K2 - gate delay
	k2 := s.hw.K2.Value()
	if k2 != s.lastK2 {
		s.gateDelay = time.Duration(k2) * time.Millisecond
		s.lastK2 = k2
		s.updateUI = true
	}
}

func (s *TGDState) handleDinEvent(isRise bool) {
	if isRise {
		now := time.Now()
		if !s.lastTrigger.IsZero() {
			oldPeriod := s.dinPeriod
			oldHz := s.dinHz
			s.dinPeriod = now.Sub(s.lastTrigger)
			calcHz(s)
			if s.dinPeriod != oldPeriod || s.dinHz != oldHz {
				s.updateUI = true
			}
		}
		s.lastTrigger = now

		delay := s.gateDelay
		if s.gateIsHigh {
			s.gateOff()
			s.scheduler.RemoveTask("gateOff") // Cancel existing gateOff
			delay = max(s.gateDelay, s.afterOffSettling)
		}
		s.scheduler.RemoveTask("gateOn") // Cancel any pending gateOn
		if s.gateRunning {
			s.scheduler.AddTask(s.gateOn, delay, "gateOn")
			// s.gateOn() // HACK: Call directly for immediate effect
		}
	} else {
		s.dinPulseWidth = time.Since(s.lastTrigger)
		// s.gateOff() // HACK: Call directly for immediate effect
	}
}


//	func (s *TGDState) gateOn() {
//		s.hw.CV1.On()
//		s.gateIsHigh = true
//		s.scheduler.AddTask(s.gateOff, s.gatePulseWidth, "gateOff")
//	}
func (s *TGDState) gateOn() {
	if s.debug {
		println("  GATE ON")
	}
	now := time.Now()

	// Check for double-trigger anomaly
	if !s.prevGateOffTime.IsZero() && now.Sub(s.prevGateOffTime) < s.gatePulseWidth {
		fmt.Printf("[WARN] Gate retriggered %dms after gate off\n", now.Sub(s.prevGateOffTime).Milliseconds())
	}

	// Check if gate triggered too late after the lastTrigger, considering gateDelay
	if !s.lastTrigger.IsZero() {
		lateBy := now.Sub(s.lastTrigger) - s.gateDelay
		if lateBy > 5*time.Millisecond {
			fmt.Printf("[ERROR] Gate triggered too late after last trigger: by %dms (gateDelay=%dms)\n", lateBy.Milliseconds(), s.gateDelay.Milliseconds())
		}
	}

	// fmt.Printf("[INFO] GATE ON  at %v (delay=%v, width=%v)\n", now, s.gateDelay, s.gatePulseWidth)
	s.prevGateOnTime = now
	s.gateIsHigh = true

	// s.scheduler.RemoveTask("gateOff", false)
	s.hw.CV1.On()

	// For diagnostics
	s.expectedGateMs = int(s.gatePulseWidth.Milliseconds())

	s.scheduler.AddTask(s.gateOff, s.gatePulseWidth, "gateOff")
}

//	func (s *TGDState) gateOff() {
//		s.hw.CV1.Off()
//		s.gateIsHigh = false
//	}
func (s *TGDState) gateOff() {
	if s.debug {
		println("  GATE OFF")
	}
	now := time.Now()
	duration := now.Sub(s.prevGateOnTime).Milliseconds()
	s.prevGateOffTime = now

	// fmt.Printf("[INFO] GATE OFF at %v (actual=%dms, expected=%dms)\n",
	// 	now, duration, s.expectedGateMs)

	if duration > int64(s.expectedGateMs)+2 {
		fmt.Printf("[ERROR] Gate lasted too long: %dms > expected %dms\n", duration, s.expectedGateMs)
	}
	if duration < int64(s.expectedGateMs)-2 {
		fmt.Printf("[ERROR] Gate too short: %dms < expected %dms\n", duration, s.expectedGateMs)
	}

	s.hw.CV1.Off()
	s.gateIsHigh = false
}

func (s *TGDState) drawScreen() {
	if s.debug {
		println("  Drawing screen...")
	}
	if s.updateUI {
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
	// s.fakeWork() // Stress the system pretending display update is heavy

	s.scheduler.AddTask(s.drawScreen, 150*time.Millisecond, "low_drawScreen")
}

func (s *TGDState) fakeWork() {
	time.Sleep(500 * time.Millisecond) // Simulate some work
	// simulate heavy work with for range loop 500000
	for i := 0; i < 500; i++ {
		// Simulate some work
		// inner loop to simulate work
		_ = i // Use i to prevent compiler optimization
		for j := 0; j < 50000000; j++ {
			for k := 0; k < 1000000; k++ {
				_ = k + j + i // Use all variables to prevent compiler optimization
			}
		}
		// print(".") // Print a dot to indicate work is being done
	}
}

func (s *TGDState) saveState() {
	if s.debug {
		println("  Saving state...")
	}
	s.scheduler.AddTask(s.saveState, 5*time.Second, "saveState")
}

func buildRange() []int {
	res := []int{}
	for i := 1; i <= 200; i++ {
		res = append(res, i)
	}
	for i := 201; i <= 500; i += 5 {
		res = append(res, i)
	}
	for i := 501; i <= 1600; i += 20 {
		res = append(res, i)
	}
	return res
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func calcHz(state *TGDState) {
	ms := state.dinPeriod.Milliseconds()
	newHz := 0.0
	if ms > 0 {
		newHz = 1000.0 / float64(ms)
	} else {
		newHz = 0
	}
	// Exponential smoothing
	const alpha = 0.2
	if state.dinHz == 0 {
		state.dinHz = newHz
	} else {
		state.dinHz = alpha*newHz + (1-alpha)*state.dinHz
	}
}
