package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenhuang/task-runner/internal/style"
)

func (m Model) updateAdd(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.input.Blur()
			m.mode = modeList
			return m, nil
		case "enter":
			title := strings.TrimSpace(m.input.Value())
			if title != "" {
				if err := m.store.AddTask(title); err != nil {
					m.err = err
				} else {
					m.cursor = len(m.store.Tasks) - 1
				}
			}
			m.input.Blur()
			m.mode = modeList
			return m, nil
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
	b.WriteString("\n")
	b.WriteString(style.Help.Render("enter: save  esc: cancel"))
	return b.String()
}
