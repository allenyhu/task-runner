package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenhuang/task-runner/internal/store"
	"github.com/allenhuang/task-runner/internal/style"
)

func (m Model) updateHistory(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "g" {
			if m.pendingG {
				m.histCursor = 0
				m.pendingG = false
				m.histMoveArmed = true
			} else {
				m.pendingG = true
			}
			return m, nil
		}
		m.pendingG = false

		switch key {
		case "esc", "q":
			if key == "q" {
				return m, tea.Quit
			}
			m.mode = modeList
			return m, nil
		case "G":
			if pw, ok := m.store.PastWeekAt(m.histWeek); ok && len(pw.Tasks) > 0 {
				m.histCursor = len(pw.Tasks) - 1
			}
			m.histMoveArmed = true
		case "h":
			// Older week (further back in PastWeeks).
			if m.histWeek < m.store.PastWeekCount()-1 {
				m.histWeek++
				m.histCursor = 0
			}
			m.histMoveArmed = true
		case "l":
			// Newer week (toward present).
			if m.histWeek > 0 {
				m.histWeek--
				m.histCursor = 0
			}
			m.histMoveArmed = true
		case "up", "k":
			if m.histCursor > 0 {
				m.histCursor--
			}
			m.histMoveArmed = true
		case "down", "j":
			if pw, ok := m.store.PastWeekAt(m.histWeek); ok && m.histCursor < len(pw.Tasks)-1 {
				m.histCursor++
			}
			m.histMoveArmed = true
		case "m":
			if !m.histMoveArmed {
				return m, nil
			}
			if pw, ok := m.store.PastWeekAt(m.histWeek); ok && len(pw.Tasks) > 0 {
				if err := m.store.MoveToCurrent(m.histWeek, m.histCursor); err != nil {
					m.err = err
				} else {
					m.clampHistoryCursor()
					m.histMoveArmed = false
				}
			}
		case "d":
			if pw, ok := m.store.PastWeekAt(m.histWeek); ok && len(pw.Tasks) > 0 {
				if err := m.store.DeletePastTask(m.histWeek, m.histCursor); err != nil {
					m.err = err
				} else {
					m.clampHistoryCursor()
				}
			}
		}
	}
	return m, nil
}

func (m *Model) clampHistoryCursor() {
	n := m.store.PastWeekCount()
	if n == 0 {
		m.histWeek = 0
		m.histCursor = 0
		return
	}
	if m.histWeek >= n {
		m.histWeek = n - 1
	}
	pw, ok := m.store.PastWeekAt(m.histWeek)
	if !ok || len(pw.Tasks) == 0 {
		m.histCursor = 0
		return
	}
	if m.histCursor >= len(pw.Tasks) {
		m.histCursor = len(pw.Tasks) - 1
	}
}

func (m Model) viewHistory() string {
	var b strings.Builder
	b.WriteString(style.Title.Render("Previous Weeks"))
	b.WriteString("\n")

	n := m.store.PastWeekCount()
	if n == 0 {
		b.WriteString(style.Empty.Render("No leftover tasks from previous weeks."))
		b.WriteString("\n")
		b.WriteString(style.Help.Render("esc: back to list"))
		return b.String()
	}

	pw, ok := m.store.PastWeekAt(m.histWeek)
	if !ok {
		b.WriteString(style.Empty.Render("No leftover tasks from previous weeks."))
		b.WriteString("\n")
		b.WriteString(style.Help.Render("esc: back to list"))
		return b.String()
	}

	start := store.WeekStart(pw.WeekStart)
	end := store.WeekEnd(start)
	b.WriteString(style.StatLabel.Render(fmt.Sprintf(
		"%s – %s  (%d/%d)",
		start.Format("Mon Jan 2"),
		end.Format("Mon Jan 2"),
		m.histWeek+1,
		n,
	)))
	b.WriteString("\n\n")

	if len(pw.Tasks) == 0 {
		b.WriteString(style.Empty.Render("No leftover tasks."))
		b.WriteString("\n")
	} else {
		for i, t := range pw.Tasks {
			plain := fmt.Sprintf("P%d  %s", store.NormalizePriority(t.Priority), t.Title)
			var rendered string
			if i == m.histCursor {
				rendered = style.Selected.Render("> " + plain)
			} else {
				rendered = style.Normal.Render("  ") +
					priorityLabel(t.Priority) +
					style.Normal.Render("  "+t.Title)
			}
			b.WriteString(rendered)
			b.WriteString("\n")
		}
	}

	b.WriteString(style.Help.Render("h/l: week  j/k: move  gg/G: top/bottom  m: to this week  d: delete  esc: back"))
	return b.String()
}
