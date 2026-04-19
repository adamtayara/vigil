//go:build darwin

package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

var launchDirs = []string{
	"/Library/LaunchDaemons",
	"/Library/LaunchAgents",
	"/System/Library/LaunchDaemons",
}

func scanTasksPlatform() ([]Task, error) {
	home, _ := os.UserHomeDir()
	dirs := append(launchDirs, filepath.Join(home, "Library/LaunchAgents"))

	var tasks []Task
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".plist") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			task, err := parsePlist(path)
			if err != nil {
				continue
			}
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func parsePlist(path string) (Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Task{}, err
	}

	var obj map[string]interface{}
	_, err = plist.Unmarshal(data, &obj)
	if err != nil {
		return Task{}, err
	}

	task := Task{Source: path}
	if label, ok := obj["Label"].(string); ok {
		task.Name = label
	}
	if prog, ok := obj["Program"].(string); ok {
		task.Action = prog
	} else if args, ok := obj["ProgramArguments"].([]interface{}); ok && len(args) > 0 {
		if s, ok := args[0].(string); ok {
			task.Action = s
		}
	}
	return task, nil
}
