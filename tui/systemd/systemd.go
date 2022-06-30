// model for a screen that watches services
package systemd

import (
	"fmt"
	"log"
	"spirit-box/services"
	g "spirit-box/tui/globals"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lp "github.com/charmbracelet/lipgloss"
	"github.com/coreos/go-systemd/v22/dbus"
	//	"log"
)

var readyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("10"))
var notReadyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("9"))
var alignRightStyle = lp.NewStyle().Align(lp.Right)
var alignLeftStyle = lp.NewStyle().Align(lp.Left)

const systemdInterval = 500 // time between updates in milliseconds

type Model struct {
	watcher     *services.UnitWatcher
	unitInfo    unitInfo
	curScreen   g.Screen
	cursorIndex int
}

func New(dConn *dbus.Conn) Model {
	watcher := services.NewWatcher(dConn)
	return Model{watcher: watcher, curScreen: g.Systemd, cursorIndex: 0}
}

func (m Model) UpdateCmd() tea.Cmd {
	return func() tea.Msg {
		m.watcher.UpdateAll()
		time.Sleep(systemdInterval * time.Millisecond)
		return g.SystemdUpdateMsg(struct{}{})
	}
}

func (m Model) Init() tea.Cmd {
	return m.UpdateCmd()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.curScreen {
	case g.Systemd:
		switch msg := msg.(type) {
		case g.SystemdUpdateMsg:
			return m, m.UpdateCmd()
		case tea.KeyMsg:
			switch msg.String() {
			case "j":
				if m.cursorIndex < len(m.watcher.Units)-1 {
					m.cursorIndex++
				}
			case "k":
				if m.cursorIndex > 0 {
					m.cursorIndex--
				}
			case "enter":
				m.unitInfo = InitUnitInfo(m.watcher.DConn, m.watcher.Units[m.cursorIndex].Name)
				return m, func() tea.Msg { return g.SwitchScreenMsg(g.UnitInfoScreen) }
			case "ctrl+c":
				return m, tea.Quit
			case "q":
				return m, func() tea.Msg { return g.SwitchScreenMsg(g.TopLevel) }
			}
		}
	case g.UnitInfoScreen:
		m.unitInfo, cmd = m.unitInfo.Update(msg)
	}

	switch msg := msg.(type) {
	case g.SwitchScreenMsg:
		m.curScreen = g.Screen(msg)
		log.Printf("From systemd, SwitchScreenMsg: %s", m.curScreen.String())
	}

	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder
	switch m.curScreen {
	case g.Systemd:
		fmt.Fprintf(&b, "Watching %d services (%.0fs):\n\n", m.watcher.NumUnits(), m.watcher.Elapsed().Seconds())

		var readyStatus string
		for i, u := range m.watcher.Units {
			if u.SubStateDesired == "watch" {
				readyStatus = readyStyle.Render("WATCHING")
			} else if u.Ready {
				readyStatus = readyStyle.Render("READY")
			} else {
				readyStatus = notReadyStyle.Render("NOT READY")
			}
			left := u.Name + ":"
			right := fmt.Sprintf("%s %s %s %s",
				alignRight(len("not-found"), u.LoadState),
				alignRight(len("activating"), u.ActiveState),
				alignRight(len("running"), u.SubState),
				alignRight(len("NOT READY"), readyStatus),
			)

			if i == m.cursorIndex {
				fmt.Fprintf(&b, "-> ")
			}

			fmt.Fprintf(&b, "%s%s %s\n", left, alignRight(80-len(left), right), u.Description)
		}

		return b.String()
	case g.UnitInfoScreen:
		return m.unitInfo.View()
	}
	return "Something went wrong!"
}

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}
