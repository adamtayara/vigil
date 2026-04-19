package scanner

import (
	"fmt"
	"strings"
	"time"

	gops "github.com/shirou/gopsutil/v4/process"
)

func ScanProcesses() ([]ProcessInfo, error) {
	procs, err := gops.Processes()
	if err != nil {
		return nil, fmt.Errorf("listing processes: %w", err)
	}

	pidMap := make(map[int32]*ProcessInfo, len(procs))
	var all []ProcessInfo

	for _, p := range procs {
		info := ProcessInfo{}
		info.PID = p.Pid

		if name, err := p.Name(); err == nil {
			info.Name = name
		}
		if exe, err := p.Exe(); err == nil {
			info.Exe = exe
		}
		if cmd, err := p.Cmdline(); err == nil {
			info.CmdLine = cmd
		}
		if ppid, err := p.Ppid(); err == nil {
			info.PPID = ppid
		}
		if user, err := p.Username(); err == nil {
			info.Username = user
		}
		if mem, err := p.MemoryInfo(); err == nil && mem != nil {
			info.MemMB = float64(mem.RSS) / 1024 / 1024
		}
		if cpu, err := p.CPUPercent(); err == nil {
			info.CPUPercent = cpu
		}
		if status, err := p.Status(); err == nil && len(status) > 0 {
			info.Status = strings.Join(status, ",")
		}
		if ct, err := p.CreateTime(); err == nil {
			info.CreateTime = time.UnixMilli(ct)
		}

		all = append(all, info)
		pidMap[info.PID] = &all[len(all)-1]
	}

	return all, nil
}

func BuildProcessTree(procs []ProcessInfo) []*ProcessInfo {
	byPID := make(map[int32]*ProcessInfo, len(procs))
	for i := range procs {
		byPID[procs[i].PID] = &procs[i]
	}

	var roots []*ProcessInfo
	for i := range procs {
		p := &procs[i]
		if parent, ok := byPID[p.PPID]; ok && p.PPID != p.PID {
			parent.Children = append(parent.Children, p)
		} else {
			roots = append(roots, p)
		}
	}
	return roots
}
