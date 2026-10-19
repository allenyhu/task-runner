package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenhuang/task-runner/internal/style"
)

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.store.Tasks)-1 {
				m.cursor++
			}
		case " ":
			if len(m.store.Tasks) > 0 {
				if err := m.store.ToggleDone(m.cursor); err != nil {
					m.err = err
				}
			}
		case "a":
			m.mode = modeAdd
			m.input.SetValue("")
			m.input.Focus()
			return m, textinput.Blink
		case "d":
			if len(m.store.Tasks) > 0 {
				if err := m.store.DeleteTask(m.cursor); err != nil {
					m.err = err
				} else if m.cursor >= len(m.store.Tasks) && m.cursor > 0 {
					m.cursor--
				}
			}
		case "r":
			m.mode = modeReview
		}
	}
	return m, nil
}

func (m Model) viewList() string {
	var b strings.Builder
	b.WriteString(style.Title.Render("Weekly Tasks"))
	b.WriteString("\n")

	if len(m.store.Tasks) == 0 {
		b.WriteString(style.Empty.Render("No tasks yet. Press a to add one."))
		b.WriteString("\n")
	} else {
		for i, t := range m.store.Tasks {
			check := "[ ]"
			if t.Done {
				check = "[x]"
			}
			line := fmt.Sprintf("%s %s", check, t.Title)

			var rendered string
			switch {
			case i == m.cursor:
				rendered = style.Selected.Render("> " + line)
			case t.Done:
				rendered = style.Done.Render("  " + line)
			default:
				rendered = style.Normal.Render("  " + line)
			}
			b.WriteString(rendered)
			b.WriteString("\n")
		}
	}

	b.WriteString(style.Help.Render("j/k: move  space: toggle  a: add  d: delete  r: review  q: quit"))
	return b.String()
}
