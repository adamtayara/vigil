//go:build linux

package scanner

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func scanTasksPlatform() ([]Task, error) {
	var tasks []Task

	// User crontab
	if out, err := exec.Command("crontab", "-l").Output(); err == nil {
		tasks = append(tasks, parseCrontab(string(out), "user crontab")...)
	}

	// System cron dirs
	cronDirs := []string{"/etc/cron.d", "/etc/cron.daily", "/etc/cron.hourly", "/etc/cron.weekly", "/etc/cron.monthly"}
	for _, dir := range cronDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			path := filepath.Join(dir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			tasks = append(tasks, parseCrontab(string(data), path)...)
		}
	}

	// Systemd timers
	if out, err := exec.Command("systemctl", "list-timers", "--all", "--no-pager").Output(); err == nil {
		tasks = append(tasks, parseSystemdTimers(string(out))...)
	}

	return tasks, nil
}

func parseCrontab(content, source string) []Task {
	var tasks []Task
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		trigger := strings.Join(fields[:5], " ")
		action := strings.Join(fields[5:], " ")
		tasks = append(tasks, Task{
			Name:    action,
			Trigger: trigger,
			Action:  action,
			Source:  source,
		})
	}
	return tasks
}

func parseSystemdTimers(output string) []Task {
	var tasks []Task
	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		name := fields[len(fields)-1]
		if strings.HasSuffix(name, ".timer") {
			tasks = append(tasks, Task{
				Name:   strings.TrimSuffix(name, ".timer"),
				Source: "systemd",
			})
		}
	}
	return tasks
}
