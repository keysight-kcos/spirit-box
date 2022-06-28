package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/dbus"
	"fmt"
	"strings"
	"os"
)

type Screen int

const (
	TopLevel Screen = iota
	Services
	Scripts
)

type switchScreenMsg Screen

type model struct {
	options []string
	cursorIndex int
	curScreen Screen
	services  serviceModel
}

func (m model) Init() tea.Cmd {
	return m.services.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.curScreen {
	case TopLevel:
		switch msg := msg.(type) {
			case tea.KeyMsg:
			switch msg.String() {
				case "j":
					if m.cursorIndex < len(m.options) - 1 {
						m.cursorIndex++
					}
				case "k":
					if m.cursorIndex > 0 {
						m.cursorIndex--
					}
				case "enter":
					if m.cursorIndex == 0 {
						m.curScreen = Services
					}
				case "q":
					return m, tea.Quit
				case "ctrl+c":
					return m, tea.Quit
			}

		}
	//case Services:
	}

	switch msg := msg.(type) {
		case switchScreenMsg:
			m.curScreen = Screen(msg)
	}

	var cmd tea.Cmd
	m.services, cmd = m.services.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.curScreen {
	case TopLevel:
		var b strings.Builder
		fmt.Fprintf(&b, "spirit-box\n\n")
		for i, option := range m.options {
			if i == m.cursorIndex {
				fmt.Fprintf(&b, "-> ")
			}
			fmt.Fprintf(&b, "%s\n", option)
		}
		return b.String()
	case Services:
		return m.services.View()
	}
	return "Something went wrong!"
}

func initialModel(dConn *dbus.Conn) model {
	return model{
		options: []string{"systemd", "scripts"},
		cursorIndex: 0,
		curScreen: TopLevel,
		services: newServiceModel(dConn),
	}
}

func StartTUI(dConn *dbus.Conn) {
	model := initialModel(dConn)
	if err := tea.NewProgram(model).Start(); err != nil {
		fmt.Printf("There was an error: %v\n", err)
		os.Exit(1)
	}
}
