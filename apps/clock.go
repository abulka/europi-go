package apps

import (
	"europi/buttons"
	"europi/controls"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"
)

// MultiPulseSync is a clock-synchronized pulse generator.
// It can sync to an external DIN signal or run on its own internal clock set by K2.
// It uses a hybrid timing system:
// - Divisions (CV1, CV2) and the 1:1 clock (CV3) are driven by a counter that increments on each DIN pulse or internal "tick".
// - Multiplications (CV4, CV5, CV6) use a time-based phase accumulator that is hard-synced by DIN/tick events.
type MultiPulseSync struct{}

func (MultiPulseSync) Name() string { return "Pulse Sync" }

// PulseOutput represents a single CV output channel.
type PulseOutput struct {
	CV      controls.ICV
	Mult    float64   // Clock multiplier. >1.0 for faster, <1.0 for slower.
	Divisor int       // For divisions (Mult < 1.0), stores the division factor (e.g., 4 for 1/4 speed).
	Phase   float64   // For multiplications (Mult > 1.0), current phase from 0.0 to 1.0.
	IsHigh  bool      // True if the CV output is currently high.
	NextOff time.Time // The time at which the current pulse should end.
}

// PulseState holds the complete state of the application.
type PulseState struct {
	hw        *controls.Controls
	btnMgr    *buttons.ButtonManager
	running   bool
	syncToDIN bool // True to sync to DIN, false for internal clock.
	cvEnabled bool // True if CV outputs are active.

	// --- Timing & Synchronization ---
	dinPeriod    time.Duration // The master clock period, set by DIN or knob.
	lastDinTime  time.Time     // Time of the last DIN signal, for period calculation.
	lastTickTime time.Time     // Time of the last processing loop tick, for calculating delta time.
	dinHz        float64
	dinCounter   int       // A counter that increments on each DIN pulse or internal tick, used for divisions.
	freeRunPhase float64   // Phase accumulator for the internal clock in free mode.
	edgeEvents   chan bool // Channel to receive rising edge events from the DIN ISR.

	knob1, knob2 int
	updateUI     bool // Flag to trigger a screen redraw.

	pulseWidth time.Duration
	pulses     []*PulseOutput
	debug      bool // If true, enables debug output to the console.

	justSwitchedToDIN bool // True if we just switched to DIN mode and need to set dinPeriod directly
}

// resetAndFireMultipliers resets phase to zero and fires a pulse for all multipliers (CV4, CV5, CV6).
// Returns a map of which pulses were just synced (by index).
func resetAndFireMultipliers(state *PulseState, now time.Time, debugLabel string) map[*PulseOutput]bool {
	justSynced := make(map[*PulseOutput]bool)
	for i, p := range state.pulses {
		if p.Mult > 1.0 {
			oldPhase := p.Phase
			p.Phase = 0.0
			if state.cvEnabled {
				firePulse(p, now, state.pulseWidth)
			}
			if state.debug {
				fmt.Printf("[%s] CV%d phase reset from %.4f to 0.0000 and pulse fired\n", debugLabel, i+1, oldPhase)
			}
			justSynced[p] = true
		}
	}
	return justSynced
}

// Run is the main entry point and loop for the application.
func (MultiPulseSync) Run(hw *controls.Controls) {
	state := &PulseState{
		hw:           hw,
		btnMgr:       buttons.New(hw.B1, hw.B2),
		running:      true,
		syncToDIN:    true,
		cvEnabled:    true,
		dinPeriod:    500 * time.Millisecond, // Default to 120 BPM.
		pulseWidth:   20 * time.Millisecond,  // A reasonable default pulse width.
		edgeEvents:   make(chan bool, 16),    // Buffered channel for ISR events.
		updateUI:     true,                   // Initial UI draw.
		dinCounter:   0,
		freeRunPhase: 0.0,
		debug:        false, // Set to true for debug output.
	}

	// Define pulse multipliers and calculate divisors where necessary.
	state.pulses = []*PulseOutput{
		{CV: hw.CV1, Mult: 0.25}, // 1/4 speed
		{CV: hw.CV2, Mult: 0.5},  // 1/2 speed
		{CV: hw.CV3, Mult: 1.0},  // 1:1 sync with DIN
		{CV: hw.CV4, Mult: 2.0},  // 2x speed
		{CV: hw.CV5, Mult: 3.0},  // 3x speed
		{CV: hw.CV6, Mult: 4.0},  // 4x speed
	}
	for _, p := range state.pulses {
		if p.Mult > 0 && p.Mult < 1.0 {
			p.Divisor = int(math.Round(1.0 / p.Mult))
		}
	}

	// --- Setup ISR for DIN sync ---
	hw.DIN.SetEdgeHandlers(
		func() { // Rising edge handler
			if state.syncToDIN {
				select {
				case state.edgeEvents <- true:
				default: // Channel is full, event is dropped. This is OK.
				}
			}
		},
		func() {}, // Falling edge is ignored.
	)

	var wg sync.WaitGroup
	defer func() {
		hw.DIN.UnsetInterrupt()
		wg.Wait()
		for _, p := range state.pulses {
			p.CV.Off()
		}
	}()

	// --- UI Update Loop ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		for state.running {
			if state.updateUI {
				state.drawScreen()
				state.updateUI = false
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// --- Main Application Loop ---
	state.lastTickTime = time.Now()
	for state.running {
		now := time.Now()
		deltaTime := now.Sub(state.lastTickTime)
		state.lastTickTime = now

		// 1. Handle knobs and buttons
		state.handleControls()

		// 2. Set tempo source from knob if in free mode
		if !state.syncToDIN {
			tempoHz := mapKnobToHz(state.knob2)
			if tempoHz > 0 {
				state.dinPeriod = time.Duration(1e9 / tempoHz)
				state.dinHz = tempoHz
			}
		}

		// 3. Check for triggers from DIN or internal free-run clock
		dinTrigger := false
		select {
		case <-state.edgeEvents:
			dinTrigger = true
		default:
		}

		freeTrigger := false
		if !state.syncToDIN && state.dinPeriod > 0 {
			increment := deltaTime.Seconds() / state.dinPeriod.Seconds()
			state.freeRunPhase += increment
			if state.freeRunPhase >= 1.0 {
				freeTrigger = true
				state.freeRunPhase -= 1.0
			}
		}

		// 4. Process Triggers
		var justSynced map[*PulseOutput]bool
		if dinTrigger {
			// --- A real DIN event occurred ---
			now := time.Now()
			if !state.lastDinTime.IsZero() {
				newPeriod := now.Sub(state.lastDinTime)
				if state.justSwitchedToDIN {
					state.dinPeriod = newPeriod
					state.justSwitchedToDIN = false
				} else {
					// More stable smoothing to reduce jitter from the source clock.
					state.dinPeriod = (state.dinPeriod*3 + newPeriod) / 4
				}
				state.dinHz = 1.0 / state.dinPeriod.Seconds()
			}
			if state.debug {
				fmt.Printf("[DIN] t=%v period=%v\n", now.Format("15:04:05.000"), state.dinPeriod)
				for i, p := range state.pulses {
					if p.Mult > 1.0 {
						fmt.Printf("[PRE-SYNC] CV%d phase=%.4f\n", i+1, p.Phase)
					}
				}
			}
			state.lastDinTime = now
			state.dinCounter++
			state.freeRunPhase = 0.0 // Reset free-run phase to sync it.

			state.processTick(now) // Fire divisions and 1:1 clock

			// Hard-sync multipliers by resetting phase to zero and firing a pulse
			justSynced = resetAndFireMultipliers(state, now, "SYNC")
			print("D")
			state.updateUI = true

		} else if freeTrigger {
			// --- A virtual free-run tick occurred ---
			now := time.Now()
			state.dinCounter++
			state.processTick(now) // Fire divisions and 1:1 clock

			// Hard-sync multipliers in free mode by resetting phase to zero and firing a pulse
			justSynced = resetAndFireMultipliers(state, now, "SYNC-FREE")
			print(".")
			state.updateUI = true
		}

		// 5. Update multiplier phases and manage all pulse-off events
		if state.cvEnabled && state.dinPeriod > 0 {
			periodSeconds := state.dinPeriod.Seconds()
			deltaSeconds := deltaTime.Seconds()

			for _, pulse := range state.pulses {
				// Turn off any pulse that has finished its duration.
				if pulse.IsHigh && now.After(pulse.NextOff) {
					pulse.CV.Off()
					pulse.IsHigh = false
				}

				// Phase accumulation logic ONLY applies to multipliers.
				if pulse.Mult > 1.0 {
					// Suppress phase accumulation and [FIRE] debug if just synced this tick
					if (dinTrigger || freeTrigger) && justSynced != nil && justSynced[pulse] {
						continue
					}
					phaseIncrement := (deltaSeconds / periodSeconds) * pulse.Mult
					pulse.Phase += phaseIncrement
					if pulse.Phase >= 1.0 {
						if state.debug {
							fmt.Printf("[FIRE] CV%d t=%v phase=%.4f\n", findCVIndex(state.pulses, pulse)+1, now.Format("15:04:05.000"), pulse.Phase)
						}
						firePulse(pulse, now, state.pulseWidth)
						pulse.Phase -= 1.0 // Wrap phase
					}
				}
			}
		}

		runtime.Gosched()
	}
}

// processTick handles the firing of counter-based pulses (divisions and 1:1).
// This is called on both real DIN events and virtual free-run ticks.
func (s *PulseState) processTick(now time.Time) {
	if !s.cvEnabled {
		return
	}
	for _, pulse := range s.pulses {
		// Divisions are triggered by the DIN counter.
		if pulse.Mult < 1.0 {
			if pulse.Divisor > 0 && (s.dinCounter-1)%pulse.Divisor == 0 {
				firePulse(pulse, now, s.pulseWidth)
			}
		} else if pulse.Mult == 1.0 {
			// The 1x pulse fires on every tick.
			firePulse(pulse, now, s.pulseWidth)
		}
	}
}

// firePulse is a helper to turn a CV on and set its off time.
func firePulse(p *PulseOutput, now time.Time, width time.Duration) {
	if !p.IsHigh {
		p.CV.On()
		p.IsHigh = true
		p.NextOff = now.Add(width)
	}
}

// handleControls processes button presses and knob turns.
func (s *PulseState) handleControls() {
	switch s.btnMgr.Update() {
	case buttons.B1Press:
		s.cvEnabled = !s.cvEnabled
		if !s.cvEnabled {
			for _, p := range s.pulses {
				if p.IsHigh {
					p.CV.Off()
					p.IsHigh = false
				}
				p.Phase = 0.0
			}
		}
		s.updateUI = true
	case buttons.B2Press:
		prevSyncToDIN := s.syncToDIN
		s.syncToDIN = !s.syncToDIN
		s.dinCounter = 0 // Reset counter when changing mode
		s.freeRunPhase = 0.0
		if !prevSyncToDIN && s.syncToDIN {
			s.justSwitchedToDIN = true
			s.lastDinTime = time.Time{} // Reset to zero so first DIN tick is ignored
		}
		s.updateUI = true
	}

	if s.btnMgr.BothHeld() {
		s.running = false
		return
	}

	k1 := s.hw.K1.Value()
	k2 := s.hw.K2.Value()
	if k1 != s.knob1 || k2 != s.knob2 {
		s.knob1, s.knob2 = k1, k2
		s.updateUI = true
	}
}

// drawScreen updates the OLED display with the current state.
func (s *PulseState) drawScreen() {
	s.hw.Display.ClearBuffer()
	syncStatus := "Free"
	if s.syncToDIN {
		syncStatus = "DIN"
	}
	cvEnabledStatus := "Off"
	if s.cvEnabled {
		cvEnabledStatus = "On"
	}

	s.hw.Display.WriteLine(0, fmt.Sprintf("Sync:%s CVs:%s", syncStatus, cvEnabledStatus))

	// K1 is displayed but currently has no function.
	s.hw.Display.WriteLine(1, fmt.Sprintf("K1: %-3d (n/a)", s.knob1))

	if s.syncToDIN {
		s.hw.Display.WriteLine(2, fmt.Sprintf("%.1fHz %dms", s.dinHz, s.dinPeriod.Milliseconds()))
	} else {
		bpm := int(s.dinHz * 60.0)
		s.hw.Display.WriteLine(2, fmt.Sprintf("K2: %d BPM", bpm))
	}

	s.hw.Display.Display()
}

// mapKnobToHz converts the K2 knob value (0-100) to a frequency in Hz.
func mapKnobToHz(knobVal int) float64 {
	// Maps knob to 40-240 BPM range.
	bpm := 40.0 + float64(knobVal)*2.0
	return bpm / 60.0 // Convert BPM to Hz.
}
