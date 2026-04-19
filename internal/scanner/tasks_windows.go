//go:build windows

package scanner

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
	"time"
)

func scanTasksPlatform() ([]Task, error) {
	cmd := exec.Command("schtasks", "/query", "/fo", "LIST", "/v")
	out, err := cmd.Output()
	if err != nil {
		// schtasks can fail with non-zero on some systems; try without /v
		out, err = exec.Command("schtasks", "/query", "/fo", "LIST").Output()
		if err != nil {
			return nil, err
		}
	}
	return parseSchTasks(out), nil
}

func parseSchTasks(data []byte) []Task {
	var tasks []Task
	var current Task
	inTask := false

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "TaskName:") {
			if inTask && current.Name != "" {
				tasks = append(tasks, current)
			}
			current = Task{}
			inTask = true
			current.Name = strings.TrimSpace(strings.TrimPrefix(line, "TaskName:"))
			// Remove leading backslash from task path
			if idx := strings.LastIndex(current.Name, "\\"); idx >= 0 {
				current.Name = current.Name[idx+1:]
			}
			continue
		}

		if !inTask {
			continue
		}

		kv := splitKV(line)
		if kv == nil {
			continue
		}
		key, val := kv[0], kv[1]

		switch key {
		case "Status":
			current.Status = val
		case "Run As User":
			current.RunAs = val
		case "Task To Run":
			current.Action = val
		case "Scheduled Task State":
			if current.Status == "" {
				current.Status = val
			}
		case "Trigger":
			if current.Trigger == "" {
				current.Trigger = val
			}
		case "Last Run Time":
			current.LastRun = parseWinTime(val)
		case "Next Run Time":
			current.NextRun = parseWinTime(val)
		}
	}
	if inTask && current.Name != "" {
		tasks = append(tasks, current)
	}
	return tasks
}

func splitKV(line string) []string {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return nil
	}
	return []string{
		strings.TrimSpace(line[:idx]),
		strings.TrimSpace(line[idx+1:]),
	}
}

func parseWinTime(s string) time.Time {
	s = strings.TrimSpace(s)
	formats := []string{
		"1/2/2006 3:04:05 PM",
		"1/2/2006 15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}
