package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenhuang/task-runner/internal/store"
	"github.com/allenhuang/task-runner/internal/style"
)

func (m Model) updateAdd(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.err = nil
			m.input.Blur()
			m.mode = modeList
			return m, nil
		case "enter":
			title := strings.TrimSpace(m.input.Value())
			if title == "" {
				m.err = nil
				m.input.Blur()
				m.mode = modeList
				return m, nil
			}
			idx, err := m.store.AddTask(title, store.PriorityP2)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.err = nil
			if idx >= 0 {
				m.cursor = idx
			}
			m.input.Blur()
			m.mode = modeList
			return m, nil
		default:
			// Typing clears a prior block error.
			m.err = nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) viewAdd() string {
	var b strings.Builder
	b.WriteString(style.Title.Render("Add Task"))
	b.WriteString("\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")
	b.WriteString(style.Help.Render("enter: save  esc: cancel"))
	return b.String()
}

func priorityLabel(p int) string {
	p = store.NormalizePriority(p)
	label := fmt.Sprintf("P%d", p)
	switch p {
	case store.PriorityP1:
		return style.PriorityP1.Render(label)
	case store.PriorityP2:
		return style.PriorityP2.Render(label)
	default:
		return style.PriorityP3.Render(label)
	}
}
