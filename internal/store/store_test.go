package store

import (
	"errors"
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

func TestSortTasksDoneAtBottomByPriority(t *testing.T) {
	s := &Store{Tasks: []Task{
		{ID: "done-p1", Title: "done high", Priority: PriorityP1, Done: true, CreatedAt: time.Unix(1, 0)},
		{ID: "open-p3", Title: "open low", Priority: PriorityP3, Done: false, CreatedAt: time.Unix(2, 0)},
		{ID: "done-p3", Title: "done low", Priority: PriorityP3, Done: true, CreatedAt: time.Unix(3, 0)},
		{ID: "open-p1", Title: "open high", Priority: PriorityP1, Done: false, CreatedAt: time.Unix(4, 0)},
		{ID: "done-p2", Title: "done mid", Priority: PriorityP2, Done: true, CreatedAt: time.Unix(5, 0)},
	}}
	s.sortTasks()
	want := []string{"open-p1", "open-p3", "done-p1", "done-p2", "done-p3"}
	for i, id := range want {
		if s.Tasks[i].ID != id {
			t.Fatalf("index %d = %s, want %s (order: %+v)", i, s.Tasks[i].ID, id, ids(s))
		}
	}
}

func TestToggleDoneMovesToBottom(t *testing.T) {
	useTempPath(t)
	s := &Store{WeekStart: WeekStart(time.Now()), Tasks: []Task{
		{ID: "a", Title: "high", Priority: PriorityP1, CreatedAt: time.Unix(1, 0)},
		{ID: "b", Title: "mid", Priority: PriorityP2, CreatedAt: time.Unix(2, 0)},
		{ID: "c", Title: "low", Priority: PriorityP3, CreatedAt: time.Unix(3, 0)},
	}}

	// Complete P1 — should move below open tasks, first among done (by priority).
	newIdx, err := s.ToggleDone(0)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 2 {
		t.Fatalf("completed P1 index = %d, want 2", newIdx)
	}
	if got := titles(s); got[0] != "mid" || got[1] != "low" || got[2] != "high" {
		t.Fatalf("order after completing high = %v, want [mid low high]", got)
	}
	if !s.Tasks[2].Done || s.Tasks[2].DoneAt == nil {
		t.Fatalf("expected high marked done with DoneAt, got %+v", s.Tasks[2])
	}

	// Complete P3 — done section ordered by priority: high (P1) then low (P3).
	newIdx, err = s.ToggleDone(1)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 2 {
		t.Fatalf("completed P3 index = %d, want 2", newIdx)
	}
	if got := titles(s); got[0] != "mid" || got[1] != "high" || got[2] != "low" {
		t.Fatalf("order after completing low = %v, want [mid high low]", got)
	}

	// Uncomplete high — returns to open section, above mid by priority.
	newIdx, err = s.ToggleDone(1)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 0 {
		t.Fatalf("uncompleted high index = %d, want 0", newIdx)
	}
	if got := titles(s); got[0] != "high" || got[1] != "mid" || got[2] != "low" {
		t.Fatalf("order after uncompleting high = %v, want [high mid low]", got)
	}
	if s.Tasks[0].Done || s.Tasks[0].DoneAt != nil {
		t.Fatalf("expected high open again, got %+v", s.Tasks[0])
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

	// Demote "high" (index 0) to P2 so the P1 slot is free, then promote "low".
	newIdx, err := s.SetPriority(0, PriorityP2)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 0 {
		t.Fatalf("demoted high index = %d, want 0 (still before mid by priority)", newIdx)
	}
	if got := titles(s); got[0] != "high" || got[1] != "mid" || got[2] != "low" {
		t.Fatalf("order after demoting high to P2 = %v, want [high mid low]", got)
	}

	// Promote "low" (index 2) to P1 — becomes sole active P1 at top.
	newIdx, err = s.SetPriority(2, PriorityP1)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 0 {
		t.Fatalf("promoted index = %d, want 0", newIdx)
	}
	if s.Tasks[newIdx].ID != "c" || s.Tasks[newIdx].Priority != PriorityP1 {
		t.Fatalf("expected c at returned index, got %+v", s.Tasks[newIdx])
	}
	if got := titles(s); got[0] != "low" || got[1] != "high" || got[2] != "mid" {
		t.Fatalf("order after promote = %v, want [low high mid]", got)
	}

	// Demote "mid" (index 2) to P3 — stays at bottom.
	newIdx, err = s.SetPriority(2, PriorityP3)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 2 {
		t.Fatalf("demoted index = %d, want 2", newIdx)
	}
	if got := titles(s); got[0] != "low" || got[1] != "high" || got[2] != "mid" {
		t.Fatalf("order after demote = %v, want [low high mid]", got)
	}
}

func TestLoadRolloverArchivesIncomplete(t *testing.T) {
	useTempPath(t)
	lastWeek := WeekStart(time.Now()).AddDate(0, 0, -7)
	s := &Store{
		WeekStart: lastWeek,
		Tasks: []Task{
			{ID: "open", Title: "leftover", Priority: PriorityP1, Done: false, CreatedAt: time.Unix(1, 0)},
			{ID: "done", Title: "finished", Priority: PriorityP2, Done: true, CreatedAt: time.Unix(2, 0)},
		},
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	thisWeek := WeekStart(time.Now())
	if !loaded.WeekStart.Equal(thisWeek) {
		t.Fatalf("WeekStart = %v, want %v", loaded.WeekStart, thisWeek)
	}
	if len(loaded.Tasks) != 0 {
		t.Fatalf("current tasks = %v, want empty", titles(loaded))
	}
	if len(loaded.PastWeeks) != 1 {
		t.Fatalf("PastWeeks len = %d, want 1", len(loaded.PastWeeks))
	}
	pw := loaded.PastWeeks[0]
	if !WeekStart(pw.WeekStart).Equal(lastWeek) {
		t.Fatalf("past WeekStart = %v, want %v", pw.WeekStart, lastWeek)
	}
	if len(pw.Tasks) != 1 || pw.Tasks[0].ID != "open" {
		t.Fatalf("past tasks = %+v, want [open]", pw.Tasks)
	}
}

func TestLoadRolloverAllDoneNoPastWeek(t *testing.T) {
	useTempPath(t)
	lastWeek := WeekStart(time.Now()).AddDate(0, 0, -7)
	s := &Store{
		WeekStart: lastWeek,
		Tasks: []Task{
			{ID: "done", Title: "finished", Priority: PriorityP2, Done: true, CreatedAt: time.Unix(1, 0)},
		},
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Tasks) != 0 {
		t.Fatalf("current tasks = %v, want empty", titles(loaded))
	}
	if len(loaded.PastWeeks) != 0 {
		t.Fatalf("PastWeeks = %+v, want empty", loaded.PastWeeks)
	}
}

func TestLoadLegacyJSONNoPastWeeks(t *testing.T) {
	useTempPath(t)
	thisWeek := WeekStart(time.Now())
	s := &Store{
		WeekStart: thisWeek,
		Tasks: []Task{
			{ID: "a", Title: "now", Priority: PriorityP2, CreatedAt: time.Unix(1, 0)},
		},
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Tasks) != 1 || loaded.Tasks[0].ID != "a" {
		t.Fatalf("tasks = %+v, want [a]", loaded.Tasks)
	}
	if loaded.PastWeeks == nil {
		t.Fatal("PastWeeks should be non-nil empty slice")
	}
	if len(loaded.PastWeeks) != 0 {
		t.Fatalf("PastWeeks = %+v, want empty", loaded.PastWeeks)
	}
}

func TestMoveToCurrent(t *testing.T) {
	useTempPath(t)
	created := time.Unix(10, 0)
	s := &Store{
		WeekStart: WeekStart(time.Now()),
		Tasks:     []Task{},
		PastWeeks: []PastWeek{{
			WeekStart: WeekStart(time.Now()).AddDate(0, 0, -7),
			Tasks: []Task{
				{ID: "x", Title: "carry", Priority: PriorityP1, CreatedAt: created},
			},
		}},
	}

	if err := s.MoveToCurrent(0, 0); err != nil {
		t.Fatal(err)
	}
	if len(s.PastWeeks) != 0 {
		t.Fatalf("PastWeeks = %+v, want pruned", s.PastWeeks)
	}
	if len(s.Tasks) != 1 {
		t.Fatalf("tasks len = %d, want 1", len(s.Tasks))
	}
	got := s.Tasks[0]
	if got.ID != "x" || got.Title != "carry" || got.Priority != PriorityP1 {
		t.Fatalf("moved task = %+v", got)
	}
	if !got.CreatedAt.Equal(created) {
		t.Fatalf("CreatedAt = %v, want %v", got.CreatedAt, created)
	}
	if got.Done || got.DoneAt != nil {
		t.Fatalf("expected incomplete, got %+v", got)
	}
}

func TestDeletePastTask(t *testing.T) {
	useTempPath(t)
	s := &Store{
		WeekStart: WeekStart(time.Now()),
		PastWeeks: []PastWeek{{
			WeekStart: WeekStart(time.Now()).AddDate(0, 0, -7),
			Tasks: []Task{
				{ID: "a", Title: "keep", Priority: PriorityP2, CreatedAt: time.Unix(1, 0)},
				{ID: "b", Title: "drop", Priority: PriorityP3, CreatedAt: time.Unix(2, 0)},
			},
		}},
	}

	if err := s.DeletePastTask(0, 1); err != nil {
		t.Fatal(err)
	}
	if len(s.PastWeeks) != 1 || len(s.PastWeeks[0].Tasks) != 1 || s.PastWeeks[0].Tasks[0].ID != "a" {
		t.Fatalf("after delete: %+v", s.PastWeeks)
	}

	if err := s.DeletePastTask(0, 0); err != nil {
		t.Fatal(err)
	}
	if len(s.PastWeeks) != 0 {
		t.Fatalf("PastWeeks = %+v, want pruned", s.PastWeeks)
	}
}

func TestBlockSecondActiveP1AddAndSet(t *testing.T) {
	useTempPath(t)
	s := &Store{WeekStart: WeekStart(time.Now()), Tasks: []Task{
		{ID: "a", Title: "focus", Priority: PriorityP1, CreatedAt: time.Unix(1, 0)},
		{ID: "b", Title: "other", Priority: PriorityP2, CreatedAt: time.Unix(2, 0)},
	}}

	if _, err := s.AddTask("another", PriorityP1); !errors.Is(err, ErrActiveP1Exists) {
		t.Fatalf("AddTask P1 err = %v, want ErrActiveP1Exists", err)
	}
	if len(s.Tasks) != 2 {
		t.Fatalf("tasks len after blocked add = %d, want 2", len(s.Tasks))
	}

	if _, err := s.SetPriority(1, PriorityP1); !errors.Is(err, ErrActiveP1Exists) {
		t.Fatalf("SetPriority P1 err = %v, want ErrActiveP1Exists", err)
	}
	if s.Tasks[1].Priority != PriorityP2 {
		t.Fatalf("priority after blocked set = %d, want P2", s.Tasks[1].Priority)
	}
}

func TestActiveP1AllowedWhenExistingParked(t *testing.T) {
	useTempPath(t)
	s := &Store{WeekStart: WeekStart(time.Now()), Tasks: []Task{
		{ID: "a", Title: "parked", Priority: PriorityP1, Parked: true, CreatedAt: time.Unix(1, 0)},
		{ID: "b", Title: "other", Priority: PriorityP2, CreatedAt: time.Unix(2, 0)},
	}}

	idx, err := s.AddTask("focus", PriorityP1)
	if err != nil {
		t.Fatal(err)
	}
	if idx != 0 {
		t.Fatalf("new P1 index = %d, want 0", idx)
	}
	if s.Tasks[0].Title != "focus" || s.Tasks[0].Parked {
		t.Fatalf("expected active focus P1 at top, got %+v", s.Tasks[0])
	}

	// Reset: only parked P1 + P3, promote P3 to P1.
	s.Tasks = []Task{
		{ID: "a", Title: "parked", Priority: PriorityP1, Parked: true, CreatedAt: time.Unix(1, 0)},
		{ID: "c", Title: "promote", Priority: PriorityP3, CreatedAt: time.Unix(3, 0)},
	}
	newIdx, err := s.SetPriority(1, PriorityP1)
	if err != nil {
		t.Fatal(err)
	}
	if newIdx != 0 || s.Tasks[0].ID != "c" {
		t.Fatalf("promoted = index %d %+v, want 0 c", newIdx, s.Tasks[newIdx])
	}
}

func TestToggleParkedFreesAndBlocksUnpark(t *testing.T) {
	useTempPath(t)
	s := &Store{WeekStart: WeekStart(time.Now()), Tasks: []Task{
		{ID: "a", Title: "focus", Priority: PriorityP1, CreatedAt: time.Unix(1, 0)},
		{ID: "b", Title: "waiting", Priority: PriorityP1, Parked: true, CreatedAt: time.Unix(2, 0)},
	}}

	idx, err := s.ToggleParked(0)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Tasks[idx].Parked {
		t.Fatalf("expected focus parked, got %+v", s.Tasks[idx])
	}
	if s.ActiveP1Index() >= 0 {
		t.Fatalf("expected no active P1 after parking focus")
	}

	// Unpark waiting — becomes sole active P1.
	waitIdx := s.IndexByID("b")
	idx, err = s.ToggleParked(waitIdx)
	if err != nil {
		t.Fatal(err)
	}
	if s.Tasks[idx].Parked || s.Tasks[idx].ID != "b" {
		t.Fatalf("expected waiting active, got %+v", s.Tasks[idx])
	}

	// Unpark focus while waiting is active — blocked.
	focusIdx := s.IndexByID("a")
	if _, err := s.ToggleParked(focusIdx); !errors.Is(err, ErrActiveP1Exists) {
		t.Fatalf("unpark err = %v, want ErrActiveP1Exists", err)
	}
	if !s.Tasks[s.IndexByID("a")].Parked {
		t.Fatal("focus should still be parked after blocked unpark")
	}
}

func TestSortActiveBeforeParked(t *testing.T) {
	s := &Store{Tasks: []Task{
		{ID: "parked-p1", Title: "parked high", Priority: PriorityP1, Parked: true, CreatedAt: time.Unix(1, 0)},
		{ID: "active-p2", Title: "active mid", Priority: PriorityP2, CreatedAt: time.Unix(2, 0)},
		{ID: "active-p1", Title: "active high", Priority: PriorityP1, CreatedAt: time.Unix(3, 0)},
		{ID: "parked-p2", Title: "parked mid", Priority: PriorityP2, Parked: true, CreatedAt: time.Unix(4, 0)},
	}}
	s.sortTasks()
	want := []string{"active-p1", "active-p2", "parked-p1", "parked-p2"}
	for i, id := range want {
		if s.Tasks[i].ID != id {
			t.Fatalf("index %d = %s, want %s (order: %+v)", i, s.Tasks[i].ID, id, ids(s))
		}
	}
}

func TestNormalizeParksExtraActiveP1s(t *testing.T) {
	s := &Store{Tasks: []Task{
		{ID: "newer", Title: "newer", Priority: PriorityP1, CreatedAt: time.Unix(2, 0)},
		{ID: "older", Title: "older", Priority: PriorityP1, CreatedAt: time.Unix(1, 0)},
	}}
	changed := s.normalizeAndSort()
	if !changed {
		t.Fatal("expected normalize to change store")
	}
	older := s.Tasks[s.IndexByID("older")]
	newer := s.Tasks[s.IndexByID("newer")]
	if older.Parked {
		t.Fatalf("oldest active P1 should stay unparked, got %+v", older)
	}
	if !newer.Parked {
		t.Fatalf("newer P1 should be parked, got %+v", newer)
	}
	if s.ActiveP1Index() != s.IndexByID("older") {
		t.Fatalf("active P1 = %d, want older", s.ActiveP1Index())
	}
}

func TestMoveToCurrentParksConflictingP1(t *testing.T) {
	useTempPath(t)
	s := &Store{
		WeekStart: WeekStart(time.Now()),
		Tasks: []Task{
			{ID: "focus", Title: "focus", Priority: PriorityP1, CreatedAt: time.Unix(1, 0)},
		},
		PastWeeks: []PastWeek{{
			WeekStart: WeekStart(time.Now()).AddDate(0, 0, -7),
			Tasks: []Task{
				{ID: "old", Title: "old p1", Priority: PriorityP1, CreatedAt: time.Unix(10, 0)},
			},
		}},
	}

	if err := s.MoveToCurrent(0, 0); err != nil {
		t.Fatal(err)
	}
	moved := s.Tasks[s.IndexByID("old")]
	if !moved.Parked || moved.Priority != PriorityP1 {
		t.Fatalf("moved task = %+v, want parked P1", moved)
	}
	if !s.Tasks[s.IndexByID("focus")].IsActiveP1() {
		t.Fatal("focus should remain the active P1")
	}
}

func TestToggleDoneClearsParked(t *testing.T) {
	useTempPath(t)
	s := &Store{WeekStart: WeekStart(time.Now()), Tasks: []Task{
		{ID: "a", Title: "parked", Priority: PriorityP2, Parked: true, CreatedAt: time.Unix(1, 0)},
	}}
	if _, err := s.ToggleDone(0); err != nil {
		t.Fatal(err)
	}
	if !s.Tasks[0].Done || s.Tasks[0].Parked {
		t.Fatalf("expected done and unparked, got %+v", s.Tasks[0])
	}
}

func titles(s *Store) []string {
	out := make([]string, len(s.Tasks))
	for i, t := range s.Tasks {
		out[i] = t.Title
	}
	return out
}

func ids(s *Store) []string {
	out := make([]string, len(s.Tasks))
	for i, t := range s.Tasks {
		out[i] = t.ID
	}
	return out
}
