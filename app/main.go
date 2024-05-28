package main

import (
	tui "docmon/models"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(
		&tui.MainModel{},
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
