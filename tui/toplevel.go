package tui

import (
	"fmt"
	"log"
	"strings"

	"spirit-box/services"
	g "spirit-box/tui/globals"
	"spirit-box/tui/scripts"
	"spirit-box/tui/systemd"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lp "github.com/charmbracelet/lipgloss"
	"github.com/coreos/go-systemd/v22/dbus"
)

var readyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("10"))
var notReadyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("9"))

type model struct {
	options     []string
	cursorIndex int
	curScreen   g.Screen
	systemd     systemd.Model
	scripts     scripts.Model
	ipStr       string
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
				} else if m.cursorIndex == 1 {
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
		var info string
		fmt.Fprintf(&b, "spirit-box\n")
		if m.systemd.AllReady {
			info = readyStyle.Render("All systemd units are ready.")
		} else {
			info = notReadyStyle.Render("Waiting for systemd units to be ready.")
		}
		fmt.Fprintf(&b, info)

		fmt.Fprintf(&b, fmt.Sprintf("\n\n%s\n\n", m.ipStr))
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

func initialModel(dConn *dbus.Conn, watcher *services.UnitWatcher, ip string) model {
	return model{
		options:     []string{"systemd", "scripts"},
		cursorIndex: 0,
		curScreen:   g.TopLevel,
		systemd:     systemd.New(dConn, watcher),
		scripts:     scripts.New(),
		ipStr:       fmt.Sprintf("Serving web ui at http://%s:8080", ip),
	}
}

func CreateProgram(dConn *dbus.Conn, watcher *services.UnitWatcher, ip string) *tea.Program {
	model := initialModel(dConn, watcher, ip)
	p := tea.NewProgram(model)
	return p
}
