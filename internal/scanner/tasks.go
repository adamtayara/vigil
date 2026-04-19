package scanner

// ScanTasks returns scheduled tasks/cron jobs.
// Platform-specific in tasks_windows.go, tasks_darwin.go, tasks_linux.go
func ScanTasks() ([]Task, error) {
	return scanTasksPlatform()
}
