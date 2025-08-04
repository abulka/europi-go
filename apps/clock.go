package apps

import (
	"europi/buttons"
	"europi/controls"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type MultiPulseSync struct{}

func (MultiPulseSync) Name() string { return "Pulse Sync" }

type PulseOutput struct {
	CV        controls.ICV
	Mult      float64 // >1 = faster, <1 = slower
	IsHigh    bool
	NextOn    time.Time
	NextOff   time.Time
}

type PulseState struct {
	hw        *controls.Controls
	btnMgr    *buttons.ButtonManager
	running   bool
	syncToDIN bool
	cvEnabled bool

	dinLastTrigger time.Time
	dinPeriod      time.Duration
	dinHz          float64
	edgeEvents     chan bool

	knob1, knob2 int
	updateUI     bool

	pulseWidth time.Duration
	pulses     []*PulseOutput
}

func (MultiPulseSync) Run(hw *controls.Controls) {
	state := &PulseState{
		hw:          hw,
		btnMgr:      buttons.New(hw.B1, hw.B2),
		running:     true,
		syncToDIN:   true,
		cvEnabled:	 true,
		pulseWidth:  40 * time.Millisecond,
		edgeEvents:  make(chan bool, 8),
	}

	// Define pulse multipliers: CV3=1x, CV2=2x, CV1=4x, CV4=0.5x, CV5=1/3x, CV6=1/4x
	state.pulses = []*PulseOutput{
		{CV: hw.CV1, Mult: 4},
		{CV: hw.CV2, Mult: 2},
		{CV: hw.CV3, Mult: 1},
		{CV: hw.CV4, Mult: 0.5},
		{CV: hw.CV5, Mult: 1.0 / 3.0},
		{CV: hw.CV6, Mult: 0.25},
	}

	// --- Setup ISR for DIN sync ---
	hw.DIN.SetEdgeHandlers(
		func() { // rising edge
			if state.syncToDIN {
				select {
				case state.edgeEvents <- true:
				default:
				}
			}
		},
		func() {}, // falling edge ignored
	)

	var wg sync.WaitGroup
	defer func() {
		hw.DIN.UnsetInterrupt()
		wg.Wait()
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

	// --- Main Loop ---
	for state.running {
		now := time.Now()

		// 1. Process DIN events
		select {
		case <-state.edgeEvents:
			triggerTime := now
			if !state.dinLastTrigger.IsZero() {
				state.dinPeriod = triggerTime.Sub(state.dinLastTrigger)
				state.dinHz = 1.0 / state.dinPeriod.Seconds()
			}
			state.dinLastTrigger = triggerTime
			state.updatePulseSchedule(triggerTime)
			state.updateUI = true
		default:
		}

		// 2. Fire or clear pulses
		for _, pulse := range state.pulses {
			if !pulse.IsHigh && !pulse.NextOn.IsZero() && now.After(pulse.NextOn) {
				pulse.CV.On()
				pulse.IsHigh = true
				pulse.NextOff = now.Add(state.pulseWidth)
			}
			if pulse.IsHigh && now.After(pulse.NextOff) {
				pulse.CV.Off()
				pulse.IsHigh = false
			}
		}

		// 3. Handle buttons
		switch state.btnMgr.Update() {
		case buttons.B1Press:
			state.cvEnabled = !state.cvEnabled
			if state.cvEnabled {
				for _, p := range state.pulses {
					p.CV.On()
					p.IsHigh = true
					p.NextOff = time.Time{}
				}
			} else {
				for _, p := range state.pulses {
					p.CV.Off()
					p.IsHigh = false
					p.NextOn = time.Time{}
					p.NextOff = time.Time{}
				}
			}
			state.updateUI = true
		case buttons.B2Press:
			state.syncToDIN = !state.syncToDIN
			state.updateUI = true
		}

		if state.btnMgr.BothHeld() {
			state.running = false
			continue
		}

		// 4. Handle knobs
		k1 := hw.K1.Value()
		k2 := hw.K2.Value()

		if k1 != state.knob1 || k2 != state.knob2 {
			state.knob1, state.knob2 = k1, k2
			state.updateUI = true
		}

		// 5. Free-running if not synced
		if !state.syncToDIN && (now.Sub(state.dinLastTrigger) > 2*time.Second) {
			tempoHz := mapKnobToHz(state.knob2)
			state.dinHz = tempoHz
			state.dinPeriod = time.Duration(1e9 / tempoHz)
			state.updatePulseSchedule(now)
		}

		runtime.Gosched()
	}
}

func (s *PulseState) updatePulseSchedule(base time.Time) {
	scaler := 1.0 + float64(s.knob1)/50.0 // 1.0–3.0
	for _, pulse := range s.pulses {
		pulseTime := time.Duration(float64(s.dinPeriod) / (pulse.Mult * scaler))
		pulse.NextOn = base.Add(pulseTime)
	}
}

func (s *PulseState) drawScreen() {
	s.hw.Display.ClearBuffer()
	syncStatus := "Free"
	if s.syncToDIN {
		syncStatus = "DIN"
	}
	cvEnabledStatus := " "
	if s.cvEnabled {
		cvEnabledStatus = "."
	}
	s.hw.Display.WriteLine(0, fmt.Sprintf("Mode: %s %.1fHz", syncStatus, s.dinHz))
	s.hw.Display.WriteLine(1, fmt.Sprintf("K1 DivMod %.1fx %s", 1+float64(s.knob1)/50.0, cvEnabledStatus))
	if !s.syncToDIN {
		bpm := int(mapKnobToHz(s.knob2) * 60.0)
		s.hw.Display.WriteLine(2, fmt.Sprintf("K2 Tempo: %d BPM", bpm))
	} else {
		s.hw.Display.WriteLine(2, fmt.Sprintf("DIN Pw: %dms", s.dinPeriod.Milliseconds()))
	}
	s.hw.Display.Display()
}

func mapKnobToHz(knobVal int) float64 {
	bpm := 40.0 + float64(knobVal)*2.0 // 40–240 BPM
	return bpm / 60.0                  // Hz
}

