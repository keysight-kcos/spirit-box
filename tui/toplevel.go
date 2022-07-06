package tui

import (
	"fmt"
	"log"
	"os"
	"strings"

	g "spirit-box/tui/globals"
	"spirit-box/tui/systemd"
	"spirit-box/tui/scripts"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/dbus"
)

type model struct {
	options     []string
	cursorIndex int
	curScreen   g.Screen
	systemd     systemd.Model
	scripts     scripts.Model
}

func (m model) Init() tea.Cmd {
	return m.systemd.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)

	switch m.curScreen {
	case g.TopLevel:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "j", "down":
				if m.cursorIndex < len(m.options)-1 {
					m.cursorIndex++
				}
			case "k", "up":
				if m.cursorIndex > 0 {
					m.cursorIndex--
				}
			case "enter":
				if m.cursorIndex == 0 {
					return m, func() tea.Msg { return g.SwitchScreenMsg(g.Systemd) }
				} else if m.cursorIndex == 1{
					return m, func() tea.Msg { return g.SwitchScreenMsg(g.Scripts) }
				}
			case "q":
				return m, tea.Quit
			case "ctrl+c":
				return m, tea.Quit
			}

		}
	case g.Systemd:
		m.systemd, cmd = m.systemd.Update(msg)
		cmds = append(cmds, cmd)
	case g.Scripts:
		m.scripts, cmd = m.scripts.Update(msg)
		cmds = append(cmds, cmd)
	case g.UnitInfoScreen:
		m.systemd, cmd = m.systemd.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case g.SwitchScreenMsg:
		m.curScreen = g.Screen(msg)
		log.Printf("From toplevel, SwitchScreenMsg: %s", m.curScreen.String())
		m.systemd, cmd = m.systemd.Update(msg)
		cmds = append(cmds, cmd)
	case g.SystemdUpdateMsg:
		m.systemd, cmd = m.systemd.Update(msg)
		cmds = append(cmds, cmd)
	case tea.WindowSizeMsg:
		m.systemd, cmd = m.systemd.Update(msg)
		cmds = append(cmds, cmd)
	case spinner.TickMsg:
		m.systemd, cmd = m.systemd.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.curScreen {
	case g.TopLevel:
		var b strings.Builder
		fmt.Fprintf(&b, "spirit-box\n\n")
		for i, option := range m.options {
			if i == m.cursorIndex {
				fmt.Fprintf(&b, "-> ")
			}
			fmt.Fprintf(&b, "%s\n", option)
		}
		return b.String()
	case g.Systemd:
		return m.systemd.View()
	case g.Scripts:
		return m.scripts.View()
	case g.UnitInfoScreen:
		return m.systemd.View()
	}
	return "Something went wrong!"
}

func initialModel(dConn *dbus.Conn) model {
	return model{
		options:     []string{"systemd", "scripts"},
		cursorIndex: 0,
		curScreen:   g.TopLevel,
		systemd:     systemd.New(dConn),
		scripts:     scripts.New(),
	}
}

func StartTUI(dConn *dbus.Conn) {
	model := initialModel(dConn)
	if err := tea.NewProgram(model).Start(); err != nil {
		fmt.Printf("There was an error: %v\n", err)
		os.Exit(1)
	}
}
