package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Title    lipgloss.Style
	Header   lipgloss.Style
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Current  lipgloss.Style
	Error    lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Bold(true).
			Margin(1, 0, 1, 2),

		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Margin(0, 0, 1, 2),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Background(lipgloss.Color("#262626")).
			Bold(true).
			Padding(0, 1),

		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1),

		Current: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D700")).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")),
	}
}