package style

import "github.com/charmbracelet/lipgloss"

// Soft, sunny palette — playful rather than stern.
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("218")). // pastel pink
		MarginBottom(1)

	Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("235")). // soft charcoal
		Background(lipgloss.Color("157")). // mint
		Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")) // warm cream

	Done = lipgloss.NewStyle().
		Foreground(lipgloss.Color("183")). // soft lilac
		Strikethrough(true)

	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("146")). // gentle periwinkle
		MarginTop(1)

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("209")) // coral, not alarm red

	StatLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color("146"))

	StatValue = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("221")) // sunny gold

	Empty = lipgloss.NewStyle().
		Foreground(lipgloss.Color("183")).
		Italic(true)

	PriorityP1 = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("213")) // candy pink

	PriorityP2 = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("222")) // soft peach

	PriorityP3 = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("159")) // sky mint
)
