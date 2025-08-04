package main

// go run ./cmd-demos/trigger-gate-simulator
// tinygo run --monitor ./cmd-demos/trigger-gate-simulator
// tinygo flash -target=pico --monitor ./cmd-demos/trigger-gate-simulator

import (
	"europi/util"
	"fmt"
	"sync"
	"time"
)

var tv *util.TimeVisualiser

func main() {
	time.Sleep(1 * time.Second) // Allow time for the system to settle

	var wg sync.WaitGroup
	running := true
	edgeEvents := make(chan bool, 1) // Channel for edge events
	tv = util.NewTimeVisualiser(3, 130, 200*time.Millisecond)

	defer func() {
		println("Cleaning up...")
		running = false
		close(edgeEvents)
		wg.Wait()
		println("Cleanup complete.")
	}()

	sendEvent := func(isRise bool) {
		select {
		case edgeEvents <- isRise:
			if isRise {
				tv.AddChar(0, '1')
			} else {
				tv.AddChar(0, '0')
			}
		default:
			panic("Edge event channel full, event dropped")
		}
	}

	// Simulate interrupts via a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Interrupt goroutine started")
		pulseInterval := 2000 * time.Millisecond
		lastCheck := time.Now()
		lastGateState := false
		for running {
			now := time.Now()
			if now.Sub(lastCheck) >= pulseInterval {
				lastCheck = now
				if lastGateState {
					sendEvent(false) // Falling edge
				} else {
					sendEvent(true) // Rising edge
				}
				lastGateState = !lastGateState
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Time visualiser updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for s := range tv.DisplayUpdates() {
			fmt.Println("\n" + s)
			if tv.IsFull() {
				println("Time visualiser display is full, clearing...")
				tv.Clear()
			}
		}
	}()

	// --- Gate scheduling state ---
	var nextGateOnTime time.Time
	var nextGateOffTime time.Time
	gateIsHigh := false
	gatePulseWidth := 800 * time.Millisecond // Example pulse width

	for running {
		now := time.Now()

		// 1. Process time-based gate events
		if !nextGateOnTime.IsZero() && now.After(nextGateOnTime) {
			// Gate ON
			tv.AddChar(1, '/')
			gateIsHigh = true
			nextGateOnTime = time.Time{} // clear
			nextGateOffTime = now.Add(gatePulseWidth)
		}
		if !nextGateOffTime.IsZero() && now.After(nextGateOffTime) {
			// Gate OFF
			tv.AddChar(1, '\\')
			gateIsHigh = false
			nextGateOffTime = time.Time{}
		}

		// 2. Process edge events
		select {
		case rise := <-edgeEvents:
			if rise {
				// If gate is already high, turn it off immediately (retrigger)
				if gateIsHigh {
					tv.AddChar(1, '\\')
					gateIsHigh = false
					nextGateOffTime = time.Time{}
				}
				// Schedule gate ON after 300ms
				nextGateOnTime = time.Now().Add(300 * time.Millisecond)
			}
			// (falling edge does nothing)
		default:
			// no event
		}

		// 3. Debug timer visualisation
		time.Sleep(50 * time.Millisecond)
		tv.AddChar(2, '.')
	}
}

// gateOn and gateOff are now inlined in the main loop
