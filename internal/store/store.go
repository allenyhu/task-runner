package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Task is a single weekly to-do item.
type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Done      bool       `json:"done"`
	CreatedAt time.Time  `json:"created_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

// Store holds the current week's tasks.
type Store struct {
	WeekStart time.Time `json:"week_start"` // Monday 00:00 local
	Tasks     []Task    `json:"tasks"`
}

// WeekStart returns Monday 00:00:00 local for the week containing t.
func WeekStart(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday -> 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

// WeekEnd returns Sunday 23:59:59.999... of the week starting at monday.
func WeekEnd(monday time.Time) time.Time {
	return monday.AddDate(0, 0, 7).Add(-time.Nanosecond)
}

func defaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".task-runner", "tasks.json"), nil
}

// Load reads ~/.task-runner/tasks.json. Missing file yields an empty current week.
// If the stored week is not the current week, tasks are cleared (week rollover).
func Load() (*Store, error) {
	path, err := defaultPath()
	if err != nil {
		return nil, err
	}

	thisWeek := WeekStart(time.Now())

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Store{WeekStart: thisWeek, Tasks: []Task{}}, nil
		}
		return nil, err
	}

	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	storedWeek := WeekStart(s.WeekStart)
	if !storedWeek.Equal(thisWeek) {
		s.WeekStart = thisWeek
		s.Tasks = []Task{}
		if err := s.Save(); err != nil {
			return nil, err
		}
		return &s, nil
	}

	s.WeekStart = thisWeek
	if s.Tasks == nil {
		s.Tasks = []Task{}
	}
	return &s, nil
}

// Save writes the store to ~/.task-runner/tasks.json.
func (s *Store) Save() error {
	path, err := defaultPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// AddTask appends a new incomplete task and persists.
func (s *Store) AddTask(title string) error {
	now := time.Now()
	s.Tasks = append(s.Tasks, Task{
		ID:        uuid.NewString(),
		Title:     title,
		Done:      false,
		CreatedAt: now,
	})
	return s.Save()
}

// ToggleDone flips done on the task at index and persists.
func (s *Store) ToggleDone(index int) error {
	if index < 0 || index >= len(s.Tasks) {
		return nil
	}
	t := &s.Tasks[index]
	t.Done = !t.Done
	if t.Done {
		now := time.Now()
		t.DoneAt = &now
	} else {
		t.DoneAt = nil
	}
	return s.Save()
}

// DeleteTask removes the task at index and persists.
func (s *Store) DeleteTask(index int) error {
	if index < 0 || index >= len(s.Tasks) {
		return nil
	}
	s.Tasks = append(s.Tasks[:index], s.Tasks[index+1:]...)
	return s.Save()
}

// Stats returns done count, total, and completion percent (0 if empty).
func (s *Store) Stats() (done, total int, percent float64) {
	total = len(s.Tasks)
	for _, t := range s.Tasks {
		if t.Done {
			done++
		}
	}
	if total == 0 {
		return 0, 0, 0
	}
	return done, total, float64(done) / float64(total) * 100
}
