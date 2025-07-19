package scheduler

import "time"

// SchedulerInterface defines the common interface for schedulers
// Both scheduler and scheduler2 implement this interface
type SchedulerInterface interface {
	AddTaskWithName(callback func(), delay time.Duration, name string)
	RemoveTask(name string, mustBeFound bool) bool
	PrintSchedule()
	GetScheduleShort() string // For demo purposes, prints task names only
	TaskCount() int
	Run()
	Stop()
	HasTask(name string) bool
}
