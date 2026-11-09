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

// ErrActiveP1Exists is returned when an operation would create a second active P1.
var ErrActiveP1Exists = errors.New("already have an active P1 — park it or demote first")

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
	Parked    bool       `json:"parked,omitempty"`
	Done      bool       `json:"done"`
	CreatedAt time.Time  `json:"created_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

// IsActiveP1 reports whether t is an incomplete, unparked P1.
func (t Task) IsActiveP1() bool {
	return !t.Done && !t.Parked && NormalizePriority(t.Priority) == PriorityP1
}

// PastWeek holds incomplete leftovers from a previous week.
type PastWeek struct {
	WeekStart time.Time `json:"week_start"` // Monday 00:00 local
	Tasks     []Task    `json:"tasks"`      // incomplete only
}

// Store holds the current week's tasks and archived leftovers.
type Store struct {
	WeekStart time.Time  `json:"week_start"` // Monday 00:00 local
	Tasks     []Task     `json:"tasks"`
	PastWeeks []PastWeek `json:"past_weeks,omitempty"` // newest first
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
// If the stored week is not the current week, incomplete tasks are archived into
// PastWeeks and the current list is cleared (week rollover).
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

	if s.PastWeeks == nil {
		s.PastWeeks = []PastWeek{}
	}

	storedWeek := WeekStart(s.WeekStart)
	if !storedWeek.Equal(thisWeek) {
		incomplete := incompleteTasks(s.Tasks)
		if len(incomplete) > 0 {
			s.PastWeeks = append([]PastWeek{{
				WeekStart: storedWeek,
				Tasks:     incomplete,
			}}, s.PastWeeks...)
		}
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
	changed := s.normalizeAndSort()
	if changed {
		if err := s.Save(); err != nil {
			return nil, err
		}
	}
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

// ActiveP1Index returns the index of the active (incomplete, unparked) P1, or -1.
func (s *Store) ActiveP1Index() int {
	for i, t := range s.Tasks {
		if t.IsActiveP1() {
			return i
		}
	}
	return -1
}

// wouldBeActiveP1 reports whether a task with the given fields would be an active P1.
func wouldBeActiveP1(priority int, parked, done bool) bool {
	return !done && !parked && NormalizePriority(priority) == PriorityP1
}

// AddTask appends a new incomplete task, sorts by priority, and persists.
// Returns the new task's index after sorting.
func (s *Store) AddTask(title string, priority int) (int, error) {
	priority = NormalizePriority(priority)
	if wouldBeActiveP1(priority, false, false) && s.ActiveP1Index() >= 0 {
		return -1, ErrActiveP1Exists
	}
	now := time.Now()
	id := uuid.NewString()
	s.Tasks = append(s.Tasks, Task{
		ID:        id,
		Title:     title,
		Priority:  priority,
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
	priority = NormalizePriority(priority)
	t := s.Tasks[index]
	if wouldBeActiveP1(priority, t.Parked, t.Done) {
		if active := s.ActiveP1Index(); active >= 0 && active != index {
			return index, ErrActiveP1Exists
		}
	}
	id := t.ID
	s.Tasks[index].Priority = priority
	s.sortTasks()
	if err := s.Save(); err != nil {
		return -1, err
	}
	return s.IndexByID(id), nil
}

// ToggleParked flips parked on the task at index, re-sorts, and persists.
// Unparking a P1 is blocked when another active P1 already exists.
func (s *Store) ToggleParked(index int) (int, error) {
	if index < 0 || index >= len(s.Tasks) {
		return index, nil
	}
	t := s.Tasks[index]
	becomingParked := !t.Parked
	if !becomingParked && wouldBeActiveP1(t.Priority, false, t.Done) {
		if active := s.ActiveP1Index(); active >= 0 && active != index {
			return index, ErrActiveP1Exists
		}
	}
	id := t.ID
	s.Tasks[index].Parked = becomingParked
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
		t.Parked = false
	} else {
		t.DoneAt = nil
		// If uncompleting would create a second active P1, auto-park.
		if wouldBeActiveP1(t.Priority, t.Parked, false) {
			for i, other := range s.Tasks {
				if i != index && other.IsActiveP1() {
					t.Parked = true
					break
				}
			}
		}
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

// PastWeekCount returns the number of archived past weeks.
func (s *Store) PastWeekCount() int {
	return len(s.PastWeeks)
}

// PastWeekAt returns the past week at index, or false if out of range.
func (s *Store) PastWeekAt(index int) (PastWeek, bool) {
	if index < 0 || index >= len(s.PastWeeks) {
		return PastWeek{}, false
	}
	return s.PastWeeks[index], true
}

// MoveToCurrent moves a leftover task from a past week into the current week.
// Preserves ID, title, priority, and CreatedAt. Prunes the past week if empty.
func (s *Store) MoveToCurrent(pastWeekIndex, taskIndex int) error {
	if pastWeekIndex < 0 || pastWeekIndex >= len(s.PastWeeks) {
		return nil
	}
	pw := &s.PastWeeks[pastWeekIndex]
	if taskIndex < 0 || taskIndex >= len(pw.Tasks) {
		return nil
	}
	t := pw.Tasks[taskIndex]
	t.Done = false
	t.DoneAt = nil
	if wouldBeActiveP1(t.Priority, t.Parked, false) && s.ActiveP1Index() >= 0 {
		t.Parked = true
	}
	pw.Tasks = append(pw.Tasks[:taskIndex], pw.Tasks[taskIndex+1:]...)
	if len(pw.Tasks) == 0 {
		s.PastWeeks = append(s.PastWeeks[:pastWeekIndex], s.PastWeeks[pastWeekIndex+1:]...)
	}
	s.Tasks = append(s.Tasks, t)
	s.sortTasks()
	return s.Save()
}

// DeletePastTask removes a leftover from a past week and prunes the week if empty.
func (s *Store) DeletePastTask(pastWeekIndex, taskIndex int) error {
	if pastWeekIndex < 0 || pastWeekIndex >= len(s.PastWeeks) {
		return nil
	}
	pw := &s.PastWeeks[pastWeekIndex]
	if taskIndex < 0 || taskIndex >= len(pw.Tasks) {
		return nil
	}
	pw.Tasks = append(pw.Tasks[:taskIndex], pw.Tasks[taskIndex+1:]...)
	if len(pw.Tasks) == 0 {
		s.PastWeeks = append(s.PastWeeks[:pastWeekIndex], s.PastWeeks[pastWeekIndex+1:]...)
	}
	return s.Save()
}

func incompleteTasks(tasks []Task) []Task {
	out := make([]Task, 0, len(tasks))
	for _, t := range tasks {
		if !t.Done {
			out = append(out, t)
		}
	}
	return out
}

// normalizeAndSort normalizes priorities and ensures at most one active P1
// (legacy data: oldest by CreatedAt stays active; extras are parked).
// Returns true if any task was changed.
func (s *Store) normalizeAndSort() bool {
	changed := false
	for i := range s.Tasks {
		norm := NormalizePriority(s.Tasks[i].Priority)
		if s.Tasks[i].Priority != norm {
			s.Tasks[i].Priority = norm
			changed = true
		}
	}

	var active []int
	for i, t := range s.Tasks {
		if t.IsActiveP1() {
			active = append(active, i)
		}
	}
	if len(active) > 1 {
		// Keep oldest CreatedAt active; park the rest.
		keep := active[0]
		for _, i := range active[1:] {
			if s.Tasks[i].CreatedAt.Before(s.Tasks[keep].CreatedAt) {
				keep = i
			}
		}
		for _, i := range active {
			if i == keep {
				continue
			}
			s.Tasks[i].Parked = true
			changed = true
		}
	}

	s.sortTasks()
	return changed
}

// sortTasks orders incomplete before done; among incomplete, active before
// parked; then priority ascending (P1 first), then CreatedAt ascending.
func (s *Store) sortTasks() {
	sort.SliceStable(s.Tasks, func(i, j int) bool {
		a, b := s.Tasks[i], s.Tasks[j]
		if a.Done != b.Done {
			return !a.Done // incomplete before done
		}
		if !a.Done && a.Parked != b.Parked {
			return !a.Parked // active before parked
		}
		pi := NormalizePriority(a.Priority)
		pj := NormalizePriority(b.Priority)
		if pi != pj {
			return pi < pj
		}
		return a.CreatedAt.Before(b.CreatedAt)
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
