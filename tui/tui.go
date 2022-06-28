package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	options []string
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
		case tea.KeyMsg:
		if msg.Type == tea.CtrlC {
			return m, tea.Quit
		}
		switch msg.String() {
			case "h":
				
			case "j":
			case "k":
			case "l":
			case "q":
				return m, tea.Quit
		}

	}
	return m, nil
}
