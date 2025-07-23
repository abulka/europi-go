package main

// Run with go run ./cmd/scheduler-demo

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"europi/scheduler" // Default scheduler implementation
)

// DemoApp demonstrates the scheduler functionality
// It uses the SchedulerInterface to allow switching between implementations
type DemoApp struct {
	scheduler scheduler.SchedulerInterface
	counter   int
	running   bool
}

func main() {
	flag.Parse()
	app := &DemoApp{
		scheduler: scheduler.New(),
		counter:   0,
		running:   true,
	}

	fmt.Println("Starting scheduler demo...")
	app.run()
}

func (app *DemoApp) run() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the scheduler in a separate goroutine
	go app.scheduler.Run()

	// Schedule initial tasks
	app.scheduler.AddTask(app.periodicTask, 1000*time.Millisecond, "periodic_task")
	app.scheduler.AddTask(app.oneTimeTask, 500*time.Millisecond, "one_time_task")
	app.scheduler.AddTask(app.printStatus, 2000*time.Millisecond, "print_status")
	app.scheduler.AddTask(app.cancelledTask, 3000*time.Millisecond, "cancelled_task")
	app.scheduler.AddTask(app.addMoreTasks, 4000*time.Millisecond, "add_more_tasks")

	app.scheduler.PrintSchedule()

	// Cancel the task after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Cancelling 'cancelled_task'...")
		app.scheduler.RemoveTask("cancelled_task", true)
	}()

	// Main scheduling loop - wait for interrupt signal
	for app.running {
		select {
		case <-sigChan:
			fmt.Println("\nReceived interrupt signal, shutting down...")
			app.stop()
			return
		default:
			time.Sleep(100 * time.Millisecond) // Reduce CPU usage
		}
	}
}

func (app *DemoApp) periodicTask() {
	app.counter++
	fmt.Printf("[%s] ⏱️ Periodic task executed #%d\n", time.Now().Format("15:04:05"), app.counter)

	// Reschedule BEFORE doing work to avoid race condition
	if app.counter < 10 {
		// Schedule the NEXT iteration first
		app.scheduler.AddTask(app.periodicTask, 500*time.Millisecond,
			fmt.Sprintf("periodic_task_%d", app.counter+1))

		// fmt.Printf("[%s] Scheduled next periodic task (#%d)\n", time.Now().Format("15:04:05"), app.counter+1)
		// app.scheduler.PrintSchedule()
	} else {
		fmt.Println("⏱️ Periodic task completed 10 iterations, stopping...")
	}
}

// oneTimeTask runs once and doesn't reschedule
func (app *DemoApp) oneTimeTask() {
	fmt.Printf("[%s] One-time task executed!\n", time.Now().Format("15:04:05"))
}

// printStatus shows the current scheduler state
func (app *DemoApp) printStatus() {
	fmt.Printf("[%s] 📝 Scheduler status: %d tasks pending (%s)\n",
		time.Now().Format("15:04:05"), app.scheduler.TaskCount(), app.scheduler.GetScheduleShort())
	// app.scheduler.PrintSchedule()

	// Reschedule for 3 seconds later
	app.scheduler.AddTask(app.printStatus, 3000*time.Millisecond, "print_status")
}

// cancelledTask should never run because it gets cancelled
func (app *DemoApp) cancelledTask() {
	fmt.Printf("[%s] ❌ This task should have been cancelled!\n", time.Now().Format("15:04:05"))
}

// addMoreTasks demonstrates adding multiple tasks at once
func (app *DemoApp) addMoreTasks() {
	fmt.Printf("[%s] Adding more tasks...\n", time.Now().Format("15:04:05"))

	// Add several tasks with different delays
	app.scheduler.AddTask(app.quickTask, 100*time.Millisecond, "quick_task_1")
	app.scheduler.AddTask(app.quickTask, 200*time.Millisecond, "quick_task_2")
	app.scheduler.AddTask(app.quickTask, 1300*time.Millisecond, "quick_task_3")

	// Add a task that will stop the demo after 10 seconds
	app.scheduler.AddTask(app.stopDemo, 10000*time.Millisecond, "stop_demo")

	println("👉 Added more tasks, current schedule:")
	// app.scheduler.PrintSchedule()
	if app.scheduler.TaskCount() == 0 {
		panic("No tasks scheduled! Something went wrong with adding tasks.")
	}
}

// quickTask for testing multiple rapid tasks
func (app *DemoApp) quickTask() {
	fmt.Printf("[%s] Quick task executed!\n", time.Now().Format("15:04:05"))
}

// stopDemo stops the entire demo
func (app *DemoApp) stopDemo() {
	fmt.Printf("[%s] Demo time limit reached, stopping...\n", time.Now().Format("15:04:05"))
	app.stop()
}

// stop cleanly shuts down the application
func (app *DemoApp) stop() {
	app.running = false
	app.scheduler.Stop()
	fmt.Println("Scheduler stopped, demo complete!")
}
