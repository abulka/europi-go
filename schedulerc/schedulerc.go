package schedulerc

// Scheduler using only channels, no mutexes

import (
	// "runtime"
	"fmt"
	"sync"
	"time"
)

// ScheduledTask represents a task to be executed at a specific time
type ScheduledTask struct {
	ID          string
	Callback    func()
	RunAt       time.Time
	lowPriority bool
}

// ChannelScheduler uses only channels, no mutexes
// Designed to handle recursive task scheduling (tasks that schedule other tasks)
type ChannelScheduler struct {
	schedule    chan ScheduledTask
	cancel      chan string
	taskChan    chan func()
	taskChanLow chan func() // low priority tasks
	wake        chan struct{}
	running     bool
}

func NewChannelScheduler(taskChan, taskChanLow chan func()) *ChannelScheduler {
	return &ChannelScheduler{
		schedule:    make(chan ScheduledTask, 64), // tasks
		cancel:      make(chan string, 32),        // tasks to cancel
		taskChan:    taskChan,                     // channel for executing tasks, create a goroutine to read from this & execute tasks
		taskChanLow: taskChanLow,                  // low priority tasks, create a goroutine to read from this & execute tasks
		running:     true,
		wake:        make(chan struct{}, 1), // not used in this version, but can be used to wake up the scheduler
	}
}

// AddTask schedules a task - non-blocking, safe to call from task callbacks
func (cs *ChannelScheduler) AddTask(callback func(), delay time.Duration, name string) {
	task := ScheduledTask{
		ID:          name,
		Callback:    callback,
		RunAt:       time.Now().Add(delay),
		lowPriority: len(name) > 4 && name[:4] == "low_",
	}

	// println("Scheduling task:", name, "delay:", delay.String()) // DEBUG

	select {
	case cs.schedule <- task:
		// Scheduled successfully
	default:
		// Schedule channel full - this is bad, increase buffer size
		println("[ERROR] Schedule channel full, task dropped:", name)
	}
}

// RemoveTask cancels a scheduled task - non-blocking, safe to call from task callbacks
func (cs *ChannelScheduler) RemoveTask(name string) {
	// println("Cancelling task:", name) // DEBUG

	select {
	case cs.cancel <- name:
		// Cancel request sent
	default:
		// Cancel channel full - this is bad, increase buffer size
		println("[ERROR] Cancel channel full, cancel dropped:", name)
	}
}

func (cs *ChannelScheduler) Wake() {
	select {
	case cs.wake <- struct{}{}:
		// Wake signal sent
	default:
		// Wake channel full - this is bad, increase buffer size
		println("[ERROR] Wake channel full, wake dropped")
	}
}

// Modified scheduler - no more busy waiting or arbitrary sleeps
func (cs *ChannelScheduler) Run(wg *sync.WaitGroup) {
	tasks := make(map[string]ScheduledTask)
	var timer *time.Timer       // pointer to a time.Timer object
	var timerC <-chan time.Time // receive-only channel of type time.Time
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for cs.running {
		// Calculate next wake time
		nextWake := cs.getNextWakeTime(tasks)

		// Reset or create timer
		if timer == nil {
			timer = time.NewTimer(nextWake) // timer has a channel field called C
			timerC = timer.C                // channel receives a value (the current time) when the timer expires.
		} else {
			timer.Stop()
			timer.Reset(nextWake) // Update the timer to the new duration wanted
		}

		select {
		case task, ok := <-cs.schedule:
			if !ok {
				// Schedule channel closed
				println("Schedule channel closed, exiting loop")
				return
			}
			tasks[task.ID] = task
			// println("ADD", task.ID) // DEBUG

		case cancelID, ok := <-cs.cancel:
			if !ok {
				// Cancel channel closed
				println("Cancel channel closed, exiting loop")
				return
			}
			delete(tasks, cancelID)
			// println("DEL", cancelID) // DEBUG

		case <-timerC:
			cs.processReadyTasks(tasks)

		case <-cs.wake:
			// Wake up signal received
		}
		time.Sleep(10 * time.Millisecond) // runtime.Gosched() doesn't work for TinyGo
	}
	wg.Done() // Signal that Run has completed
}

func (cs *ChannelScheduler) getNextWakeTime(tasks map[string]ScheduledTask) time.Duration {
	if len(tasks) == 0 {
		// No tasks - sleep longer but not forever
		return 1 * time.Second
	}

	now := time.Now()
	nextRun := now.Add(time.Hour) // far future default

	for _, task := range tasks {
		if task.RunAt.Before(nextRun) {
			nextRun = task.RunAt
		}
	}

	duration := nextRun.Sub(now)
	if duration <= 0 {
		return 0 // Execute immediately
	}

	// Cap the maximum sleep to handle clock issues
	// The cap may cause more frequent wakeups, but not early execution.
	cap := 1 * time.Millisecond
	if duration < cap {
		return duration
	}

	return duration
}

func (cs *ChannelScheduler) processReadyTasks(tasks map[string]ScheduledTask) {
	// println("Processing ready tasks") // DEBUG
	now := time.Now()
	for id, task := range tasks {
		if !now.Before(task.RunAt) {
			delete(tasks, id)

			// println("FAKE EXECUTE:", id) // DEBUG fake execute
			
			if task.lowPriority {
				select {
				case cs.taskChanLow <- task.Callback:
					// Task sent successfully
				default:
					println("[ERROR] processReadyTasks - low priority task dropped (channel full):", id)
				}
			} else {
				select {
				case cs.taskChan <- task.Callback:
					// Task sent successfully
				default:
					println("[ERROR] processReadyTasks - task dropped (channel full):", id)
				}
			}
		}
	}
}

func (cs *ChannelScheduler) TaskExecutor(wg *sync.WaitGroup) {
	// Task executors
	defer wg.Done()
	for task := range cs.taskChan {
		if !cs.running {
			return
		}
		task()
	}
}

func (cs *ChannelScheduler) Stop() {
	cs.running = false
	cs.Wake() // allow the goroutine to exit cleanly
	close(cs.schedule)
	close(cs.cancel)
	close(cs.wake)
}

func Demo() {
	time.Sleep(1 * time.Second) // Allow time for the app to start up and --monitor to connect
	println("Scheduler starting")
	scheduler := NewChannelScheduler(make(chan func()), make(chan func()))

	wg := sync.WaitGroup{}

	wg.Add(1)
	go scheduler.TaskExecutor(&wg)
	time.Sleep(1 * time.Second) // Allow time for the task executor to start for TinyGo

	// Example tasks
	var task2 func()
	task2NumIterations := 5
	task2 = func() {
		println(" EXECUTED Task 2")
		task2NumIterations--
		if task2NumIterations > 0 {
			scheduler.AddTask(task2, 100*time.Millisecond, fmt.Sprintf("task2 (reschedule %d)", task2NumIterations))
		}
	}
	scheduler.AddTask(func() {
		println(" EXECUTED Task 1")
	}, 2*time.Second, "task1")
	scheduler.AddTask(task2, 1*time.Second, "task2")
	scheduler.AddTask(func() {
		println(" EXECUTED Task 3")
		scheduler.RemoveTask("task1")
	}, 0*time.Second, "task3")
	scheduler.AddTask(func() {
		println(" EXECUTED Task 4")
	}, 4*time.Second, "task4")

	wg.Add(1)
	go scheduler.Run(&wg)
	time.Sleep(1 * time.Second) // Allow time for the scheduler to start for TinyGo

	time.Sleep(5 * time.Second) // Let the scheduler run for a while
	println("Scheduler stopping")
	scheduler.Stop()
	wg.Wait() // Wait for all goroutines to finish
	println("Scheduler stopped")
}
