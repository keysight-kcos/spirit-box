package tui_lite

import (
	"fmt"
	"log"
	"strings"

	"spirit-box/device"
	"spirit-box/scripts"
	"spirit-box/services"
	g "spirit-box/tui_lite/globals"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lp "github.com/charmbracelet/lipgloss"
	"github.com/coreos/go-systemd/v22/dbus"
)

var readyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("10"))
var notReadyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("9"))
var alignRightStyle = lp.NewStyle().Align(lp.Right)

const (
	width  = 500
	height = 100
	vPos   = height / 2
	hPos   = width / 2
)

type model struct {
	ipStr      string
	spinner    spinner.Model
	wipe       bool
	whitespace string
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg { return m.spinner.Tick() }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	m.wipe = false

	switch msg := msg.(type) {
	/*
		case g.RestoreScreenMsg: // closes render loop immediately after screen wipe
			log.Print("restore")
			return m, tea.Batch(cmds...)
	*/
	case g.WipeScreenMsg:
		log.Print("wipe")
		//fmt.Printf("\033[2J") // clear the screen
		m.wipe = true
		/*
			cmds = append(cmds, func() tea.Msg {
				return g.RestoreScreenMsg(struct{}{})
			})
		*/
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			cmds = append(cmds, func() tea.Msg {
				return g.WipeScreenMsg(struct{}{})
			})
			return m, tea.Batch(cmds...)
		case "q":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.wipe {
		return m.whitespace
	}

	var b strings.Builder
	//var info string
	fmt.Fprintf(&b, "spirit-box\n")

	/*
		systemdReady := m.systemd.AllReady
		scriptsReady := m.scripts.AllReady
		if systemdReady {
			info = readyStyle.Render("All systemd units are ready.")
		} else {
			info = notReadyStyle.Render("Waiting for systemd units to be ready.")
		}
		fmt.Fprintf(&b, info)

		if scriptsReady {
			info = readyStyle.Render("\nAll scripts have succeeded.")
		} else {
			info = notReadyStyle.Render("\nAll scripts have not succeeded.")
		}
		fmt.Fprintf(&b, info)

		if systemdReady && scriptsReady {
			fmt.Fprintf(&b, readyStyle.Render("\n\nSystem is ready. Press 'q' to close spirit-box."))
		}
	*/

	fmt.Fprintf(&b, fmt.Sprintf("\n\n%s\n\n", m.ipStr))

	/*
		var readyStatus string
			for _, u := range m.systemd.Watcher.Units {
				if u.Ready {
					readyStatus = readyStyle.Render("READY")
				} else {
					readyStatus = notReadyStyle.Render(m.spinner.View())
				}
				left := u.Name + ":"
				fmt.Fprintf(&b, "%s%s\n", left, alignRight(100-len(left), readyStatus))
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
				fmt.Fprintf(&b, "%s%s\n", s.Cmd, alignRight(100-len(s.Cmd), readyStatus))
			}
	*/

	fmt.Fprintf(&b, "\nPress 'r' to manually re-render the screen.\n")

	return lp.PlaceHorizontal(width, 0, b.String())
}

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}

func initialModel(dConn *dbus.Conn, watcher *services.UnitWatcher, ip string, sc *scripts.ScriptController) model {
	s := spinner.New()
	s.Spinner = spinner.Line
	whitespace := ""
	for i := 0; i < height; i++ {
		whitespace += "\n"
	}
	return model{
		ipStr:      fmt.Sprintf("Web UI at %s, ports %s, %s", ip, device.HOST_PORT, device.SERVER_PORT),
		spinner:    s,
		whitespace: whitespace,
	}
}

func CreateProgram(dConn *dbus.Conn, watcher *services.UnitWatcher, ip string, sc *scripts.ScriptController) *tea.Program {
	model := initialModel(dConn, watcher, ip, sc)
	p := tea.NewProgram(model, tea.WithAltScreen())
	// update ticker
	/*
		go func(p *tea.Program) {
			for {
				p.Send(g.WipeScreenMsg(struct{}{}))
				time.Sleep(time.Duration(1250) * time.Millisecond)
			}
		}(p)
	*/
	return p
}
