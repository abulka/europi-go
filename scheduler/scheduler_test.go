package scheduler

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_AddTaskWithName(t *testing.T) {
	s := New()
	taskName := "test_task"
	s.AddTaskWithName(func() {
		// Task logic
	}, 1000*time.Millisecond, taskName)

	if s.TaskCount() != 1 {
		t.Errorf("expected 1 task, got %d", s.TaskCount())
	}

	if !s.HasTask(taskName) {
		t.Errorf("expected task '%s' to exist", taskName)
	}
}

func TestScheduler_RemoveTask(t *testing.T) {
	s := New()
	taskName := "test_task"
	s.AddTaskWithName(func() {
		// Task logic
	}, 1000*time.Millisecond, taskName)

	s.RemoveTask(taskName, true)

	if s.TaskCount() != 0 {
		t.Errorf("expected 0 tasks, got %d", s.TaskCount())
	}

	if s.HasTask(taskName) {
		t.Errorf("expected task '%s' to be removed", taskName)
	}
}

func TestScheduler_Run(t *testing.T) {
	s := New()
	taskExecuted := false
	s.AddTaskWithName(func() {
		taskExecuted = true
	}, 100*time.Millisecond, "test_task")

	go s.Run()
	time.Sleep(200 * time.Millisecond)

	if !taskExecuted {
		t.Errorf("expected task to be executed")
	}

	s.Stop()
}

func TestScheduler_TaskExecutionOrder(t *testing.T) {
	s := New()
	taskOrder := []string{}

	s.AddTaskWithName(func() {
		taskOrder = append(taskOrder, "short_delay_task")
	}, 5*time.Millisecond, "short_delay_task")

	s.AddTaskWithName(func() {
		taskOrder = append(taskOrder, "long_delay_task")
	}, 10*time.Millisecond, "long_delay_task")

	go s.Run()
	time.Sleep(20 * time.Millisecond)

	s.Stop()

	if len(taskOrder) != 2 {
		t.Errorf("expected 2 tasks to be executed, got %d", len(taskOrder))
	}

	if taskOrder[0] != "short_delay_task" || taskOrder[1] != "long_delay_task" {
		t.Errorf("tasks executed in wrong order: %v", taskOrder)
	}
}

func TestScheduler_TaskAddsAnotherTask(t *testing.T) {
	s := New()
	taskOrder := []string{}

	s.AddTaskWithName(func() {
		taskOrder = append(taskOrder, "initial_task")
		s.AddTaskWithName(func() {
			taskOrder = append(taskOrder, "added_task")
		}, 50*time.Millisecond, "added_task")
	}, 50*time.Millisecond, "initial_task")

	go s.Run()
	time.Sleep(300 * time.Millisecond)

	s.Stop()

	if len(taskOrder) != 2 {
		t.Errorf("expected 2 tasks to be executed, got %d", len(taskOrder))
	}

	if taskOrder[0] != "initial_task" || taskOrder[1] != "added_task" {
		t.Errorf("tasks executed in wrong order: %v", taskOrder)
	}
}

func TestScheduler_PeriodicTaskReschedules(t *testing.T) {
	s := New()
	counter := 0

	var periodicTask func()
	periodicTask = func() {
		fmt.Printf("Periodic task executed #%d\n", counter)
		counter++
		if counter < 4 {
			s.AddTaskWithName(periodicTask, 50*time.Millisecond, fmt.Sprintf("periodic_task_%d", counter))
		}
	}

	s.AddTaskWithName(periodicTask, 50*time.Millisecond, "periodic_task_1")

	go s.Run()
	defer s.Stop()

	time.Sleep(300 * time.Millisecond)

	if counter != 4 {
		t.Errorf("expected counter to be 4, got %d", counter)
	}
}

func TestScheduler_RemoveRace_AddAfterRemove(t *testing.T) {
	/*
		Here’s what it effectively tests now:

		Goroutine Safety: It confirms that RemoveTask (called from a separate
		goroutine) and AddTask (called from the main test goroutine) can run
		concurrently with the scheduler's Run() loop without causing data
		corruption.

		State Consistency: It simulates a realistic sequence: a task is scheduled,
		an external event cancels it, and then another event reschedules it. The
		final assertion (executed != 1) correctly verifies that only the re-added
		task runs, and the original one (which was removed before its scheduled
		time) does not.

		Timing and Race Conditions: While the time.Sleep calls are not a perfect way
		to guarantee a race, they create a plausible scenario where the RemoveTask
		call happens while the scheduler is "between ticks" of its RunOnce loop.
		This is a crucial real-world case to validate.

		In short, keep this test. It verifies that your scheduler behaves correctly
		not just within a single, isolated RunOnce call, but as a continuously
		running service interacting with other parts of an application.
	*/
	s := New()
	executed := 0
	taskName := "animation"

	task := func() {
		executed++
		fmt.Printf("Task executed (%s): %d\n", taskName, executed)
	}

	// Add the task
	s.AddTaskWithName(task, 100*time.Millisecond, taskName)

	// Start the scheduler
	go s.Run()
	defer s.Stop()

	// In a goroutine, attempt to remove the task shortly after
	go func() {
		time.Sleep(10 * time.Millisecond) // short delay to cause a potential race
		s.RemoveTask(taskName, true)
		fmt.Println("Task removed")
	}()

	// Slight delay then re-add the task with the same name
	time.Sleep(20 * time.Millisecond)
	s.AddTaskWithName(task, 100*time.Millisecond, taskName)
	fmt.Println("Task re-added")

	// Allow scheduler time to run
	time.Sleep(300 * time.Millisecond)

	if executed != 1 {
		t.Errorf("Expected task to execute once, but got %d", executed)
	}
}

func TestRemoveTaskWhileInFirstTask(t *testing.T) {
	// Only checks the logic within a single RunOnce cycle.
	s := New()
	taskOneExecuted := false
	taskTwoExecuted := false

	s.AddTaskWithName(func() {
		taskOneExecuted = true
		s.RemoveTask("test_task2", false)
	}, 0*time.Millisecond, "test_task1")

	s.AddTaskWithName(func() {
		taskTwoExecuted = true
	}, 0*time.Millisecond, "test_task2")

	s.RunOnce()

	// Check taskOneExecuted is true, taskTwoExecuted is false and length of tasks is 0
	if !taskOneExecuted {
		t.Errorf("Expected task_one to be executed, but it was not")
	}
	if taskTwoExecuted {
		t.Errorf("Expected task_two to not be executed, but it was")
	}
	if s.TaskCount() != 0 {
		t.Errorf("Expected no tasks in the scheduler, but found %d", s.TaskCount())
	}
}

func TestScheduler_Hammer(t *testing.T) {
	/*
		The "Hammer" Test (High-Concurrency Stress Test) This test throws many
		concurrent add/remove operations at the scheduler to try and force a race
		condition or deadlock.

		Goal: To ensure the scheduler remains stable and correct under heavy,
		chaotic load.
	*/
	s := New()
	go s.Run()
	defer s.Stop()

	var wg sync.WaitGroup
	taskCount := 100
	iterations := 5

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			taskName := fmt.Sprintf("hammer_task_%d", id)
			for j := 0; j < iterations; j++ {
				// Add a task with a short, random delay
				s.AddTaskWithName(func() {}, time.Duration(rand.Intn(50))*time.Millisecond, taskName)

				// Immediately try to remove it
				s.RemoveTask(taskName, false)
			}
		}(i)
	}

	wg.Wait()
	// The main assertion is that the test completes without deadlocking or panicking.
	// You could also check that the final task count is zero.
	if s.TaskCount() != 0 {
		t.Errorf("Expected scheduler to be empty after hammer test, but found %d tasks", s.TaskCount())
	}
}

func TestScheduler_RescheduleChain(t *testing.T) {
	s := New()
	executionLimit := 5
	var executionCount int32 // Use atomic for safe concurrent access
	taskName := "chain_task"
	var wg sync.WaitGroup
	wg.Add(executionLimit)

	var taskFunc func()
	taskFunc = func() {
		atomic.AddInt32(&executionCount, 1)
		// If we haven't reached the limit, reschedule myself.
		if atomic.LoadInt32(&executionCount) < int32(executionLimit) {
			s.AddTaskWithName(taskFunc, 10*time.Millisecond, taskName) // Reschedule
		}
		wg.Done()
	}

	// 1. Start the scheduler in the background. It will now run continuously.
	go s.Run()

	// 2. Add the first task to kick off the chain.
	s.AddTaskWithName(taskFunc, 0*time.Millisecond, taskName)

	// 3. Wait here until wg.Done() has been called 5 times.
	// This is the crucial step that waits for the test condition to be met.
	wg.Wait()

	// 4. Stop the scheduler's background loop.
	s.Stop()

	// 5. Now we can safely make our assertions.
	finalCount := atomic.LoadInt32(&executionCount)
	if finalCount != int32(executionLimit) {
		t.Errorf("Expected %d executions, but got %d", executionLimit, finalCount)
	}
	if s.TaskCount() != 0 {
		t.Errorf("Expected no tasks left, but found %d", s.TaskCount())
	}
}
