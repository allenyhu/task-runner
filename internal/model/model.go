package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/allenhuang/task-runner/internal/store"
	"github.com/allenhuang/task-runner/internal/style"
)

type mode int

const (
	modeList mode = iota
	modeAdd
	modeReview
)

// Model is the root Bubble Tea model.
type Model struct {
	store  *store.Store
	mode   mode
	cursor int
	input  textinput.Model
	err    error
	width  int
	height int
}

// New creates a Model bound to an already-loaded store.
func New(s *store.Store) Model {
	ti := textinput.New()
	ti.Placeholder = "What needs doing?"
	ti.CharLimit = 200
	ti.Width = 50

	return Model{
		store: s,
		mode:  modeList,
		input: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	switch m.mode {
	case modeAdd:
		return m.updateAdd(msg)
	case modeReview:
		return m.updateReview(msg)
	default:
		return m.updateList(msg)
	}
}

func (m Model) View() string {
	var body string
	switch m.mode {
	case modeAdd:
		body = m.viewAdd()
	case modeReview:
		body = m.viewReview()
	default:
		body = m.viewList()
	}

	if m.err != nil {
		body += "\n" + style.Error.Render(fmt.Sprintf("error: %v", m.err))
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}
