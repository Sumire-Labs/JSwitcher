package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"javaswitcher/internal/ui"
)

func main() {
	model := ui.NewModel()
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}