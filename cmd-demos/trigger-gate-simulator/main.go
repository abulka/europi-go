package main

// go run ./cmd-demos/trigger-gate-simulator
// tinygo flash -target=pico --monitor ./cmd-demos/trigger-gate-simulator

import (
	"fmt"
	"sync"
	"time"
	// "runtime"
	scheduler "europi/schedulerc"
	"europi/util"
)

var sched *scheduler.ChannelScheduler
var tv *util.TimeVisualiser

func main() {
	time.Sleep(1 * time.Second) // Allow time for the system to settle

	var wg sync.WaitGroup
	running := true
	edgeEvents := make(chan bool, 1)   // Channel for edge events
	taskChan := make(chan func(), 10)
	taskChanLow := make(chan func(), 10)
	
	sched = scheduler.NewChannelScheduler(taskChan, taskChanLow)
	tv = util.NewTimeVisualiser(3, 130, 200*time.Millisecond)

	defer func() {
		println("Cleaning up...")
		running = false
		close(edgeEvents)
		close(taskChan)
		if sched != nil {
			sched.Stop()
		}
		wg.Wait()
		println("Cleanup complete.")
	}()

	// Start the scheduler in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		sched.Run(&wg)
	}()
	time.Sleep(1 * time.Second) // Allow time for the scheduler to start for TinyGo

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
				// runtime.Gosched()
			}
			// time.Sleep(500 * time.Millisecond) // extra sleep
		}		
	}()
	time.Sleep(1 * time.Second) // Allow time for the scheduler to start for TinyGo

	// Task executors
	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range taskChan {
			if !running {
				return
			}
			task()
		}
	}()
	time.Sleep(1 * time.Second) // Allow time for the scheduler to start for TinyGo

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
	time.Sleep(1 * time.Second) // Allow time for the scheduler to start for TinyGo


	// Main Loop
	var timer *time.Timer       // pointer to a time.Timer object
	var timerC <-chan time.Time // receive-only channel of type time.Time
	for {
		// Calculate next wake time - CHANGING TO 500MS DOESN'T FIX THE CRASH
		nextWake := 50 * time.Millisecond

		if timer == nil {
			timer = time.NewTimer(nextWake) // timer has a channel field called C
			timerC = timer.C                // channel receives a value (the current time) when the timer expires.
		} else {
			timer.Stop()
			timer.Reset(nextWake) // Update the timer to the new duration wanted
		}

		select {
		case rise := <-edgeEvents:
			if rise {
				// sched.RemoveTask("gateOn") // Cancel any pending gateOn ONLY DO THIS IF GATE IS ALREADY HIGH
				sched.AddTask(gateOn, 300*time.Millisecond, "gateOn")
			} else {
			}

		case <-timerC:
			tv.AddChar(2, '.') // DEBUG
		}

		time.Sleep(1 * time.Millisecond) // runtime.Gosched() doesn't work
	}

}

func gateOn() {
	// print("Gate ON\n")
	gatePulseWidth := 800 * time.Millisecond // Example pulse width
	tv.AddChar(1, '/')
	sched.AddTask(gateOff, gatePulseWidth, "gateOff") // HACK

}
func gateOff() {
	tv.AddChar(1, '\\')
}
