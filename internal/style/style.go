package style

import "github.com/charmbracelet/lipgloss"

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginBottom(1)

	Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	Done = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Strikethrough(true)

	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	StatLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	StatValue = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	Empty = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
)
