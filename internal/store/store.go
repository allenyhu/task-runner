package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
)

// Priority levels (P1 highest / top of list).
const (
	PriorityP1 = 1
	PriorityP2 = 2
	PriorityP3 = 3
)

// NormalizePriority clamps p to P1–P3; zero/invalid becomes P3.
func NormalizePriority(p int) int {
	switch p {
	case PriorityP1, PriorityP2, PriorityP3:
		return p
	default:
		return PriorityP3
	}
}

// Task is a single weekly to-do item.
type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Priority  int        `json:"priority"` // 1, 2, or 3
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

// pathFunc resolves the tasks.json path. Overridden in tests.
var pathFunc = defaultPath

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
	path, err := pathFunc()
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
	s.normalizeAndSort()
	return &s, nil
}

// Save writes the store to ~/.task-runner/tasks.json.
func (s *Store) Save() error {
	path, err := pathFunc()
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

// AddTask appends a new incomplete task, sorts by priority, and persists.
// Returns the new task's index after sorting.
func (s *Store) AddTask(title string, priority int) (int, error) {
	now := time.Now()
	id := uuid.NewString()
	s.Tasks = append(s.Tasks, Task{
		ID:        id,
		Title:     title,
		Priority:  NormalizePriority(priority),
		Done:      false,
		CreatedAt: now,
	})
	s.sortTasks()
	if err := s.Save(); err != nil {
		return -1, err
	}
	return s.IndexByID(id), nil
}

// SetPriority updates priority on the task at index, re-sorts, and persists.
// Returns the task's new index after sorting.
func (s *Store) SetPriority(index, priority int) (int, error) {
	if index < 0 || index >= len(s.Tasks) {
		return index, nil
	}
	id := s.Tasks[index].ID
	s.Tasks[index].Priority = NormalizePriority(priority)
	s.sortTasks()
	if err := s.Save(); err != nil {
		return -1, err
	}
	return s.IndexByID(id), nil
}

// ToggleDone flips done on the task at index, re-sorts, and persists.
// Returns the task's new index after sorting.
func (s *Store) ToggleDone(index int) (int, error) {
	if index < 0 || index >= len(s.Tasks) {
		return index, nil
	}
	id := s.Tasks[index].ID
	t := &s.Tasks[index]
	t.Done = !t.Done
	if t.Done {
		now := time.Now()
		t.DoneAt = &now
	} else {
		t.DoneAt = nil
	}
	s.sortTasks()
	if err := s.Save(); err != nil {
		return -1, err
	}
	return s.IndexByID(id), nil
}

// DeleteTask removes the task at index and persists.
func (s *Store) DeleteTask(index int) error {
	if index < 0 || index >= len(s.Tasks) {
		return nil
	}
	s.Tasks = append(s.Tasks[:index], s.Tasks[index+1:]...)
	return s.Save()
}

// IndexByID returns the index of the task with the given ID, or -1.
func (s *Store) IndexByID(id string) int {
	for i, t := range s.Tasks {
		if t.ID == id {
			return i
		}
	}
	return -1
}

func (s *Store) normalizeAndSort() {
	for i := range s.Tasks {
		s.Tasks[i].Priority = NormalizePriority(s.Tasks[i].Priority)
	}
	s.sortTasks()
}

// sortTasks orders incomplete tasks first, then completed; within each group
// by priority ascending (P1 first), then CreatedAt ascending.
func (s *Store) sortTasks() {
	sort.SliceStable(s.Tasks, func(i, j int) bool {
		if s.Tasks[i].Done != s.Tasks[j].Done {
			return !s.Tasks[i].Done // incomplete before done
		}
		pi := NormalizePriority(s.Tasks[i].Priority)
		pj := NormalizePriority(s.Tasks[j].Priority)
		if pi != pj {
			return pi < pj
		}
		return s.Tasks[i].CreatedAt.Before(s.Tasks[j].CreatedAt)
	})
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
