package store

import (
	"testing"
	"time"
)

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
