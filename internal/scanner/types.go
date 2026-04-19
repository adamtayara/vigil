package scanner

import "time"

type ProcessInfo struct {
	PID        int32
	PPID       int32
	Name       string
	Exe        string
	CmdLine    string
	Username   string
	CPUPercent float64
	MemMB      float64
	Status     string
	CreateTime time.Time
	Children   []*ProcessInfo
}

type Connection struct {
	PID        int32
	ProcessName string
	LocalAddr  string
	LocalPort  uint32
	RemoteAddr string
	RemotePort uint32
	Status     string
	Family     uint32
	Type       uint32
}

type DiskUsage struct {
	Path        string
	Label       string
	Total       uint64
	Used        uint64
	Free        uint64
	PercentUsed float64
	FSType      string
}

type DirSize struct {
	Path  string
	Bytes int64
}

type DiskInfo struct {
	Drives  []DiskUsage
	TopDirs []DirSize
	TopFiles []DirSize
	TempSize int64
	TempPath string
}

type Software struct {
	Name        string
	Publisher   string
	Version     string
	InstallDate time.Time
	InstallLoc  string
	Source      string
}

type Task struct {
	Name       string
	Status     string
	Action     string
	Trigger    string
	LastRun    time.Time
	NextRun    time.Time
	Created    time.Time
	RunAs      string
	Source     string
}

type Extension struct {
	Browser     string
	ID          string
	Name        string
	Version     string
	Description string
	Permissions []string
	Source      string
	UpdatedAt   time.Time
	ProfilePath string
}

type SystemInfo struct {
	Hostname     string
	OS           string
	Platform     string
	Kernel       string
	Uptime       uint64
	BootTime     time.Time
}
