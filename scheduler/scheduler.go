package scheduler

import (
	"fmt"
	"runtime"
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
	ScheduledAt  time.Time
	CallbackName string
	lowPriority  bool
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

// AddTask schedules a task with a specific name for identification
func (s *Scheduler) AddTask(callback func(), delay time.Duration, name string) {
	if s.HasTask(name) {
		s.RemoveTask(name, false)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledAt := time.Now().Add(delay)
	id := fmt.Sprintf("%s_%d", name, scheduledAt.UnixMilli())

	lowPriority := false
	if strings.HasPrefix(name, "low_") {
		lowPriority = true
		name = strings.TrimPrefix(name, "low_")
	}

	task := Task{
		ID:           id,
		Callback:     callback,
		ScheduledAt:  scheduledAt,
		CallbackName: name,
		lowPriority:  lowPriority,
	}
	s.tasks[id] = task
	// println("Task scheduled:", name, "at", scheduledAt.Format(time.RFC3339), "ID:", id)
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

func (s *Scheduler) getReadyTasks() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var ready []Task

	// Snapshot ready tasks
	for _, task := range s.tasks {
		if !task.ScheduledAt.After(now) { // Equivalent to: Before or Equal
			ready = append(ready, task)
		}
	}
	if len(ready) != 0 {
		// Optional: stable sort by ScheduledAt
		sort.SliceStable(ready, func(i, j int) bool {
			return ready[i].ScheduledAt.Before(ready[j].ScheduledAt)
		})
	}
	return ready
}

// stillExists checks if a task exists by ID and deletes it if present. Returns true if deleted.
// Assumes you are about to execute the task callback, hence needs to be removed from the schedule.
func (s *Scheduler) stillExists(id string) bool {
	s.mu.Lock()
	_, stillExists := s.tasks[id]
	if stillExists {
		delete(s.tasks, id)
	}
	s.mu.Unlock()
	return stillExists
}

// RunOnce identifies ready tasks and sends them to taskChan and taskChanLowPriority.
// Never blocks. Unlocks before task execution.
func (s *Scheduler) RunOnceTasks(taskChan, taskChanLowPriority chan<- func()) {
	if !s.enabled {
		return
	}
	ready := s.getReadyTasks()
	if len(ready) == 0 {
		return
	}

	for _, taskToRun := range ready {
		if s.stillExists(taskToRun.ID) {
			if taskToRun.lowPriority {
				select {
				case taskChanLowPriority <- taskToRun.Callback:
					// sent successfully
				default:
					// channel full, task dropped (optional: log or reschedule)
				}
			} else {
				select {
				case taskChan <- taskToRun.Callback:
					// sent successfully
				default:
					// channel full, task dropped (optional: log or reschedule)
				}
			}
		}
	}
}

// RunOnce executes all tasks that are ready to run, ensuring removed tasks are not executed.
func (s *Scheduler) RunOnce() {
	if !s.enabled {
		return
	}
	ready := s.getReadyTasks()
	if len(ready) == 0 {
		return
	}

	for _, taskToRun := range ready {
		if s.stillExists(taskToRun.ID) {
			taskToRun.Callback()
		}
	}
}

// Run starts the scheduler loop
func (s *Scheduler) Run() {
	fmt.Println("Scheduler started, running tasks...")
	for s.IsEnabled() {
		s.RunOnce()
		runtime.Gosched()
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

	now := time.Now()
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
		return tasks[i].ScheduledAt.Before(tasks[j].ScheduledAt)
	})

	for _, task := range tasks {
		timeRemaining := task.ScheduledAt.Sub(now)
		fmt.Printf("In %4dms %-15s\n", timeRemaining.Milliseconds(), task.CallbackName)
	}
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
