package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenhuang/task-runner/internal/store"
	"github.com/allenhuang/task-runner/internal/style"
)

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "g" {
			if m.pendingG {
				m.cursor = 0
				m.pendingG = false
			} else {
				m.pendingG = true
			}
			return m, nil
		}
		m.pendingG = false

		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "G":
			if n := len(m.store.Tasks); n > 0 {
				m.cursor = n - 1
			}
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
				idx, err := m.store.ToggleDone(m.cursor)
				if err != nil {
					m.err = err
				} else if idx >= 0 {
					m.cursor = idx
				}
			}
		case "1", "2", "3":
			if len(m.store.Tasks) > 0 {
				p := int(key[0] - '0')
				idx, err := m.store.SetPriority(m.cursor, p)
				if err != nil {
					m.err = err
				} else if idx >= 0 {
					m.cursor = idx
				}
			}
		case "a":
			m.mode = modeAdd
			m.addPriority = store.PriorityP2
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
		case "h":
			m.mode = modeHistory
			m.histWeek = 0
			m.histCursor = 0
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
			plain := fmt.Sprintf("%s P%d  %s", check, store.NormalizePriority(t.Priority), t.Title)

			var rendered string
			switch {
			case i == m.cursor:
				rendered = style.Selected.Render("> " + plain)
			case t.Done:
				rendered = style.Done.Render("  " + plain)
			default:
				rendered = style.Normal.Render("  "+check+" ") +
					priorityLabel(t.Priority) +
					style.Normal.Render("  "+t.Title)
			}
			b.WriteString(rendered)
			b.WriteString("\n")
		}
	}

	b.WriteString(style.Help.Render("j/k: move  gg/G: top/bottom  space: toggle  1/2/3: priority  a: add  d: delete  r: review  h: history  q: quit"))
	return b.String()
}
