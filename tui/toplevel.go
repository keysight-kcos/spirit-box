package tui

import (
	"fmt"
	"log"
	"strings"
	"time"

	"spirit-box/device"
	"spirit-box/scripts"
	"spirit-box/services"
	g "spirit-box/tui/globals"
	"spirit-box/tui/scriptsTui"
	"spirit-box/tui/systemd"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lp "github.com/charmbracelet/lipgloss"
	"github.com/coreos/go-systemd/v22/dbus"
)

var readyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("10"))
var notReadyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("9"))
var alignRightStyle = lp.NewStyle().Align(lp.Right)

type model struct {
	options     []string
	cursorIndex int
	curScreen   g.Screen
	systemd     systemd.Model
	scripts     scriptsTui.Model
	ipStr       string
	spinner     spinner.Model
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.systemd.Init(), func() tea.Msg { return m.spinner.Tick() })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

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
	case g.CheckSystemdMsg:
		m.systemd, cmd = m.systemd.Update(msg)
		cmds = append(cmds, cmd)
	case g.CheckScriptsMsg:
		m.scripts, cmd = m.scripts.Update(msg)
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

		if m.scripts.AllReady {
			info = readyStyle.Render("\nAll scripts have succeeded.")
		} else {
			info = notReadyStyle.Render("\nAll scripts have not succeeded.")
		}
		fmt.Fprintf(&b, info)

		fmt.Fprintf(&b, fmt.Sprintf("\n\n%s\n\n", m.ipStr))

		var readyStatus string
		for _, u := range m.systemd.Watcher.Units {
			if u.Ready {
				readyStatus = readyStyle.Render("READY")
			} else {
				readyStatus = notReadyStyle.Render(m.spinner.View())
			}
			left := u.Name + ":"
			fmt.Fprintf(&b, "%s%s\n", left, alignRight(80-len(left), readyStatus))
		}

		for _, s := range m.scripts.GetScriptStatuses() {
			if s.Status == 0 {
				continue
			}
			switch s.Status {
			case 2:
				readyStatus = notReadyStyle.Render("FAILED")
			case 3:
				readyStatus = readyStyle.Render("SUCCEEDED")
			default:
				readyStatus = notReadyStyle.Render(m.spinner.View())
			}
			fmt.Fprintf(&b, "%s%s\n", s.Cmd, alignRight(80-len(s.Cmd), readyStatus))
		}

		fmt.Fprintf(&b, "\n")
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

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}

func initialModel(dConn *dbus.Conn, watcher *services.UnitWatcher, ip string, sc *scripts.ScriptController) model {
	s := spinner.New()
	s.Spinner = spinner.Line
	return model{
		options:     []string{"systemd", "scripts"},
		cursorIndex: 0,
		curScreen:   g.TopLevel,
		systemd:     systemd.New(dConn, watcher),
		scripts:     scriptsTui.New(sc),
		ipStr:       fmt.Sprintf("Web UI at %s, ports %s, %s", ip, device.HOST_PORT, device.SERVER_PORT),
		spinner:     s,
	}
}

func CreateProgram(dConn *dbus.Conn, watcher *services.UnitWatcher, ip string, sc *scripts.ScriptController) *tea.Program {
	model := initialModel(dConn, watcher, ip, sc)
	p := tea.NewProgram(model, tea.WithAltScreen())
	// update ticker
	go func(p *tea.Program) {
		for {
			//p.Send(g.WipeScreenMsg(struct{}{}))
			p.Send(g.CheckSystemdMsg(struct{}{}))
			p.Send(g.CheckScriptsMsg(struct{}{}))
			time.Sleep(time.Second)
		}
	}(p)
	return p
}
