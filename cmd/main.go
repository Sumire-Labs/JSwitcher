/*
 * JSwitcher
 *
 * Copyright 2025 s12kuma01
 *
 * This software is licensed under the Open Software License version
 * 3.0. The full text of this license can be found in https://opensource.org/licenses/OSL-3.0
 * or in the LICENSES directory which is distributed along with the software.
 */

package main

import (
	"log"

	"javaswitcher/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model := ui.NewModel()
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
