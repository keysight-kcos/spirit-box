// model for a screen that watches services
package tui

import (
	"fmt"
	lp "github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/dbus"
	"spirit-box/services"
	"strings"
	"time"
//	"log"
)

var readyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("10"))
var notReadyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("9"))

const systemdInterval = 500 // time between updates in milliseconds

type systemdUpdateMsg bool

type serviceModel struct {
	watcher *services.UnitWatcher
}

func newServiceModel(dConn *dbus.Conn) serviceModel {
	watcher := services.NewWatcher(dConn)
	return serviceModel{watcher}
}

func (s serviceModel) UpdateCmd() tea.Cmd {
	return func() tea.Msg {
		s.watcher.UpdateAll()
		time.Sleep(systemdInterval*time.Millisecond)
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
	fmt.Fprintf(&b, "Watching %d services (%.4fs):\n\n", s.watcher.NumUnits(), s.watcher.Elapsed().Seconds())

	var readyStatus string
	for _, u := range s.watcher.Units {
		if u.Ready {
			readyStatus = readyStyle.Render("READY")
		} else {
			readyStatus = notReadyStyle.Render("NOT READY")
		}
		fmt.Fprintf(&b, "%s: %s %s %s %s\n", u.Name, u.LoadState, u.ActiveState, u.SubState, readyStatus)
	}

	return b.String()
}
