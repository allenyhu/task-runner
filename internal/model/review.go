package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenhuang/task-runner/internal/store"
	"github.com/allenhuang/task-runner/internal/style"
)

func (m Model) updateReview(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "r", "q":
			if msg.String() == "q" {
				return m, tea.Quit
			}
			m.mode = modeList
		}
	}
	return m, nil
}

func (m Model) viewReview() string {
	done, total, percent := m.store.Stats()
	start := m.store.WeekStart
	end := store.WeekEnd(start)

	var b strings.Builder
	b.WriteString(style.Title.Render("Weekly Review"))
	b.WriteString("\n")
	b.WriteString(style.StatLabel.Render(fmt.Sprintf(
		"Week of %s – %s",
		start.Format("Mon Jan 2"),
		end.Format("Mon Jan 2"),
	)))
	b.WriteString("\n\n")

	b.WriteString(style.StatLabel.Render("Completed  "))
	b.WriteString(style.StatValue.Render(fmt.Sprintf("%d / %d", done, total)))
	b.WriteString("\n")

	b.WriteString(style.StatLabel.Render("Rate       "))
	b.WriteString(style.StatValue.Render(fmt.Sprintf("%.0f%%", percent)))
	b.WriteString("\n")

	if total == 0 {
		b.WriteString("\n")
		b.WriteString(style.Empty.Render("No tasks this week yet."))
		b.WriteString("\n")
	}

	b.WriteString(style.Help.Render("esc/r: back to list"))
	return b.String()
}
