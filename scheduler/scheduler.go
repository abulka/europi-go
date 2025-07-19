package scheduler

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// go run ./cmd/scheduler-demo

// Task represents a scheduled task
type Task struct {
	ID           string
	Callback     func()
	ScheduledAt  int64
	CallbackName string
}

// Scheduler manages and executes scheduled tasks
// This version uses a TaskRunner-style approach with no pending task logic
// and full goroutine safety.
type Scheduler struct {
	enabled bool
	tasks   map[string]Task // for quick lookup by ID, the single source of truth
	mu      sync.Mutex
}

// Ensure Scheduler implements the updated SchedulerInterface
var _ SchedulerInterface = (*Scheduler)(nil)

// New creates a new scheduler instance
func New() *Scheduler {
	return &Scheduler{
		enabled: true,
		tasks:   make(map[string]Task),
	}
}

// AddTask schedules a task to run after the specified delay
func (s *Scheduler) AddTask(callback func(), delay time.Duration) {
	s.AddTaskWithName(callback, delay, getFunctionName(callback))
}

// AddTaskWithName schedules a task with a specific name for identification
func (s *Scheduler) AddTaskWithName(callback func(), delay time.Duration, name string) {
	if s.HasTask(name) {
		msg := fmt.Sprintf("Task with name '%s' already exists in the schedule", name)
		// remove the existing task
		s.RemoveTask(name, false)
		// double check
		if s.HasTask(name) {
			msg += " but could not be removed."
			panic(msg)
		} else {
			msg += " It has been removed and replaced with the new task."
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledAt := time.Now().Add(delay).UnixMilli()
	id := fmt.Sprintf("%s_%d", name, scheduledAt)
	task := Task{
		ID:           id,
		Callback:     callback,
		ScheduledAt:  scheduledAt,
		CallbackName: name,
	}

	s.tasks[id] = task
}

// RemoveTask removes all tasks with the specified callback name
func (s *Scheduler) RemoveTask(name string, mustBeFound bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for id, task := range s.tasks {
		if task.CallbackName == name {
			delete(s.tasks, id)
			found = true
		}
	}

	if mustBeFound && !found {
		panic(fmt.Sprintf("Task with name '%s' not found in the schedule", name))
	}
	return found
}

// RunOnce executes all tasks that are ready to run, ensuring removed tasks are not executed.
func (s *Scheduler) RunOnce() {
	s.mu.Lock()
	if !s.enabled {
		s.mu.Unlock()
		return
	}

	now := time.Now().UnixMilli()
	ready := make([]Task, 0)

	// 1. Create a snapshot of ready tasks WITHOUT removing them yet.
	for _, task := range s.tasks {
		if task.ScheduledAt <= now {
			ready = append(ready, task)
		}
	}
	s.mu.Unlock() // Unlock after creating the snapshot.

	if len(ready) == 0 {
		return
	}

	// Sort for predictable execution order.
	sort.SliceStable(ready, func(i, j int) bool {
		return ready[i].ScheduledAt < ready[j].ScheduledAt
	})

	// 2. Loop through the snapshot and re-validate each task before running.
	for _, taskToRun := range ready {
		s.mu.Lock()
		// 2a. Check if the task still exists in the main schedule.
		if _, ok := s.tasks[taskToRun.ID]; !ok {
			// It was removed by a previous task in this same batch. Skip it.
			s.mu.Unlock()
			continue
		}

		// 2b. It exists, so remove it now, right before we run it.
		delete(s.tasks, taskToRun.ID)
		s.mu.Unlock() // Unlock BEFORE running the callback to prevent deadlocks.

		// 3. Execute the callback.
		taskToRun.Callback()
	}
}

// Run starts the scheduler loop
func (s *Scheduler) Run() {
	fmt.Println("Scheduler started, running tasks...")
	for s.IsEnabled() {
		s.RunOnce()
		time.Sleep(10 * time.Millisecond)
	}
}

// Stop disables the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

// IsEnabled returns whether the scheduler is enabled
func (s *Scheduler) IsEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enabled
}

// GetScheduleShort returns a short summary of the scheduled tasks
func (s *Scheduler) GetScheduleShort() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, 0, len(s.tasks))
	for _, task := range s.tasks {
		names = append(names, task.CallbackName)
	}
	return strings.Join(names, ", ")
}

// TaskCount returns the number of scheduled tasks
func (s *Scheduler) TaskCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tasks)
}

// PrintSchedule prints all scheduled tasks for debugging
func (s *Scheduler) PrintSchedule() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	fmt.Println("Schedule:")
	if len(s.tasks) == 0 {
		fmt.Println("  No tasks in the schedule.")
		return
	}

	// Convert to slice and sort by scheduled time
	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].ScheduledAt < tasks[j].ScheduledAt
	})

	for _, task := range tasks {
		timeRemaining := task.ScheduledAt - now
		fmt.Printf("In %4dms %-15s\n", timeRemaining, task.CallbackName)
	}
}

// getFunctionName attempts to get a readable name for debugging
func getFunctionName(fn func()) string {
	return fmt.Sprintf("task_%p", fn)
}

// HasTask checks if a task with the given name is scheduled
func (s *Scheduler) HasTask(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, task := range s.tasks {
		if task.CallbackName == name {
			return true
		}
	}
	return false
}
