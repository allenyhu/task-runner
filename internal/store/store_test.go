package store

import (
	"path/filepath"
	"testing"
	"time"
)

func useTempPath(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	prev := pathFunc
	pathFunc = func() (string, error) { return path, nil }
	t.Cleanup(func() { pathFunc = prev })
}

func TestWeekStart(t *testing.T) {
	loc := time.Local
	// Wednesday Oct 15, 2025
	wed := time.Date(2025, 10, 15, 14, 30, 0, 0, loc)
	got := WeekStart(wed)
	want := time.Date(2025, 10, 13, 0, 0, 0, 0, loc) // Monday
	if !got.Equal(want) {
		t.Fatalf("WeekStart(Wed) = %v, want %v", got, want)
	}

	// Sunday Oct 19, 2025 should belong to week starting Oct 13
	sun := time.Date(2025, 10, 19, 12, 0, 0, 0, loc)
	got = WeekStart(sun)
	if !got.Equal(want) {
		t.Fatalf("WeekStart(Sun) = %v, want %v", got, want)
	}

	// Monday itself
	mon := time.Date(2025, 10, 13, 9, 0, 0, 0, loc)
	got = WeekStart(mon)
	if !got.Equal(want) {
		t.Fatalf("WeekStart(Mon) = %v, want %v", got, want)
	}
}

func TestStats(t *testing.T) {
	s := &Store{Tasks: []Task{
		{Title: "a", Done: true},
		{Title: "b", Done: false},
		{Title: "c", Done: true},
	}}
	done, total, pct := s.Stats()
	if done != 2 || total != 3 || pct < 66.6 || pct > 66.7 {
		t.Fatalf("Stats() = %d, %d, %f", done, total, pct)
	}

	empty := &Store{}
	done, total, pct = empty.Stats()
	if done != 0 || total != 0 || pct != 0 {
		t.Fatalf("empty Stats() = %d, %d, %f", done, total, pct)
	}
}

func TestNormalizePriority(t *testing.T) {
	cases := map[int]int{0: PriorityP3, 1: PriorityP1, 2: PriorityP2, 3: PriorityP3, 9: PriorityP3}
	for in, want := range cases {
		if got := NormalizePriority(in); got != want {
			t.Fatalf("NormalizePriority(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestSortTasksByPriority(t *testing.T) {
	s := &Store{Tasks: []Task{
		{ID: "c", Title: "low", Priority: PriorityP3, CreatedAt: time.Unix(3, 0)},
		{ID: "a", Title: "high", Priority: PriorityP1, CreatedAt: time.Unix(1, 0)},
		{ID: "b", Title: "mid", Priority: PriorityP2, CreatedAt: time.Unix(2, 0)},
	}}
	s.sortTasks()
	if s.Tasks[0].ID != "a" || s.Tasks[1].ID != "b" || s.Tasks[2].ID != "c" {
		t.Fatalf("unexpected order: %+v", s.Tasks)
	}
}

func TestAddTaskOrdersByPriority(t *testing.T) {
	useTempPath(t)
	s := &Store{WeekStart: WeekStart(time.Now()), Tasks: []Task{}}

	idx3, err := s.AddTask("low", PriorityP3)
	if err != nil {
		t.Fatal(err)
	}
	if idx3 != 0 {
		t.Fatalf("first task index = %d, want 0", idx3)
	}

	idx1, err := s.AddTask("high", PriorityP1)
	if err != nil {
		t.Fatal(err)
	}
	if idx1 != 0 {
		t.Fatalf("P1 insert index = %d, want 0", idx1)
	}
	if s.Tasks[0].Title != "high" || s.Tasks[1].Title != "low" {
		t.Fatalf("order after P1 insert: %+v", titles(s))
	}

	idx2, err := s.AddTask("mid", PriorityP2)
	if err != nil {
		t.Fatal(err)
	}
	if idx2 != 1 {
		t.Fatalf("P2 insert index = %d, want 1", idx2)
	}
	if got := titles(s); got[0] != "high" || got[1] != "mid" || got[2] != "low" {
		t.Fatalf("final order = %v, want [high mid low]", got)
	}
}

func TestSetPriorityReordersAndReturnsIndex(t *testing.T) {
	useTempPath(t)
	s := &Store{WeekStart: WeekStart(time.Now()), Tasks: []Task{
		{ID: "a", Title: "high", Priority: PriorityP1, CreatedAt: time.Unix(1, 0)},
		{ID: "b", Title: "mid", Priority: PriorityP2, CreatedAt: time.Unix(2, 0)},
		{ID: "c", Title: "low", Priority: PriorityP3, CreatedAt: time.Unix(3, 0)},
	}}

	// Promote "low" (index 2) to P1. Same-priority ties keep CreatedAt order,
	// so it sits after the older P1 ("high").
	newIdx, err := s.SetPriority(2, PriorityP1)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 1 {
		t.Fatalf("promoted index = %d, want 1", newIdx)
	}
	if s.Tasks[newIdx].ID != "c" || s.Tasks[newIdx].Priority != PriorityP1 {
		t.Fatalf("expected c at returned index, got %+v", s.Tasks[newIdx])
	}
	if got := titles(s); got[0] != "high" || got[1] != "low" || got[2] != "mid" {
		t.Fatalf("order after promote = %v, want [high low mid]", got)
	}

	// Demote "mid" (index 2) to P3 — only P3 left, so it stays at bottom.
	newIdx, err = s.SetPriority(2, PriorityP3)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 2 {
		t.Fatalf("demoted index = %d, want 2", newIdx)
	}
	if got := titles(s); got[0] != "high" || got[1] != "low" || got[2] != "mid" {
		t.Fatalf("order after demote = %v, want [high low mid]", got)
	}

	// Demote original "high" (index 0) to P3 — should move to bottom (CreatedAt before mid).
	newIdx, err = s.SetPriority(0, PriorityP3)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 1 {
		t.Fatalf("demoted high index = %d, want 1 (before mid by CreatedAt)", newIdx)
	}
	if got := titles(s); got[0] != "low" || got[1] != "high" || got[2] != "mid" {
		t.Fatalf("order after demoting high = %v, want [low high mid]", got)
	}
}

func titles(s *Store) []string {
	out := make([]string, len(s.Tasks))
	for i, t := range s.Tasks {
		out[i] = t.Title
	}
	return out
}
