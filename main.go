package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenhuang/task-runner/internal/model"
	"github.com/allenhuang/task-runner/internal/store"
)

func main() {
	s, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load tasks: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model.New(s), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
