package tray

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/robfig/cron/v3"
)

type Schedule struct {
	Enabled  bool   `json:"enabled"`
	CronExpr string `json:"cron"`
}

var defaultSchedules = map[string]string{
	"daily":   "0 9 * * *",
	"weekly":  "0 9 * * 1",
	"monthly": "0 9 1 * *",
}

func ScheduleExpr(name string) string {
	if expr, ok := defaultSchedules[name]; ok {
		return expr
	}
	return name
}

type Scheduler struct {
	cr       *cron.Cron
	configPath string
	schedule Schedule
}

func NewScheduler() *Scheduler {
	home, _ := os.UserHomeDir()
	return &Scheduler{
		cr:         cron.New(),
		configPath: filepath.Join(home, ".vigil", "schedule.json"),
	}
}

func (s *Scheduler) Load() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil // no config yet
	}
	return json.Unmarshal(data, &s.schedule)
}

func (s *Scheduler) Save() error {
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.schedule, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0600)
}

func (s *Scheduler) Start(scanFn func()) error {
	if err := s.Load(); err != nil {
		return err
	}
	if !s.schedule.Enabled || s.schedule.CronExpr == "" {
		return nil
	}
	_, err := s.cr.AddFunc(s.schedule.CronExpr, scanFn)
	if err != nil {
		return err
	}
	s.cr.Start()
	return nil
}

func (s *Scheduler) SetSchedule(expr string, enabled bool) error {
	s.schedule = Schedule{Enabled: enabled, CronExpr: expr}
	return s.Save()
}

func (s *Scheduler) Stop() {
	s.cr.Stop()
}
