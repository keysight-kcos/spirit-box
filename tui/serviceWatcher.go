// model for a screen that watches services
package tui

import (
	"fmt"
	"spirit-box/services"
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

type systemdUpdateMsg bool

type serviceModel struct {
	watcher      *services.UnitWatcher
	selectedUnit selectedUnit
	cursorIndex  int
}

func newServiceModel(dConn *dbus.Conn) serviceModel {
	watcher := services.NewWatcher(dConn)
	return serviceModel{watcher: watcher, cursorIndex: 0}
}

func (s serviceModel) UpdateCmd() tea.Cmd {
	return func() tea.Msg {
		s.watcher.UpdateAll()
		time.Sleep(systemdInterval * time.Millisecond)
		return systemdUpdateMsg(true)
	}
}

func (s serviceModel) Init() tea.Cmd {
	return s.UpdateCmd()
}

func (s serviceModel) Update(msg tea.Msg) (serviceModel, tea.Cmd) {
	switch msg := msg.(type) {
	case systemdUpdateMsg:
		return s, s.UpdateCmd()
	case tea.KeyMsg:
		switch msg.String() {
		case "j":
			if s.cursorIndex < len(s.watcher.Units)-1 {
				s.cursorIndex++
			}
		case "k":
			if s.cursorIndex > 0 {
				s.cursorIndex--
			}
		case "enter":
			s.selectedUnit = InitSelectedUnit(s.watcher.DConn, s.watcher.Units[s.cursorIndex].Name)
			return s, func() tea.Msg { return switchScreenMsg(UnitInfoPage) }
		case "ctrl+c":
			return s, tea.Quit
		case "q":
			return s, func() tea.Msg { return switchScreenMsg(TopLevel) }
		}
	}
	return s, nil
}

func (s serviceModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Watching %d services (%.0fs):\n\n", s.watcher.NumUnits(), s.watcher.Elapsed().Seconds())

	var readyStatus string
	for i, u := range s.watcher.Units {
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

		if i == s.cursorIndex {
			fmt.Fprintf(&b, "-> ")
		}

		fmt.Fprintf(&b, "%s%s %s\n", left, alignRight(80-len(left), right), u.Description)
	}

	return b.String()
}

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}
