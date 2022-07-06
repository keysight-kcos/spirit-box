// model for a screen that watches services
package systemd

import (
	"fmt"
	"log"
	"spirit-box/services"
	g "spirit-box/tui/globals"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	watcher           *services.UnitWatcher
	unitInfo          unitInfo
	curScreen         g.Screen
	cursorIndex       int
	spinner           spinner.Model
	textinput         textinput.Model
	textinputSelected bool
	allReady          bool
	/*
		The two fields below are used when adding new units
		while the program is running.
	*/
	addUnitBeforeUpdate bool
	newUnitName         string
	width               int
	height              int
}

func New(dConn *dbus.Conn, watcher *services.UnitWatcher) Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	t := textinput.New()
	t.Placeholder = "Press \"/\" to add more units to watch"
	return Model{
		watcher:     watcher,
		curScreen:   g.Systemd,
		cursorIndex: 0,
		spinner:     s,
		textinput:   t,
		allReady:    false,
	}
}

func (m Model) UpdateCmd() tea.Cmd {
	return func() tea.Msg {
		if m.addUnitBeforeUpdate {
			m.watcher.AddUnit(m.newUnitName)
		}
		allReady := m.watcher.UpdateAll()
		time.Sleep(systemdInterval * time.Millisecond)
		return g.SystemdUpdateMsg(allReady)
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.UpdateCmd(), func() tea.Msg { return m.spinner.Tick() })
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)
	switch m.curScreen {
	case g.Systemd:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if m.textinputSelected {
				switch msg.String() {
				case "/", "esc":
					m.textinputSelected = false
					m.textinput.Blur()
					m.textinput.SetValue("")
				case "enter":
					m.textinputSelected = false
					m.textinput.Blur()
					m.newUnitName = m.textinput.Value()
					m.addUnitBeforeUpdate = true
					m.textinput.SetValue("")
				}
			} else {
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
					m.unitInfo = InitUnitInfo(m.watcher.Units[m.cursorIndex].Properties, m.width, m.height)
					cmd := func() tea.Msg { return g.SwitchScreenMsg(g.UnitInfoScreen) }
					cmds = append(cmds, cmd)
				case "/":
					m.textinputSelected = true
					m.textinput.Focus()
					m.textinput.SetCursorMode(textinput.CursorBlink)
					return m, tea.Batch(cmds...)
				case "ctrl+c":
					return m, tea.Quit
				case "q":
					cmd := func() tea.Msg { return g.SwitchScreenMsg(g.TopLevel) }
					cmds = append(cmds, cmd)
				}
			}
		}
	case g.UnitInfoScreen:
		m.unitInfo, cmd = m.unitInfo.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case g.SystemdUpdateMsg:
		cmds = append(cmds, m.UpdateCmd())
		m.addUnitBeforeUpdate = false
		//log.Printf("From systemd, SystemddUpdateMsg")
		m.allReady = bool(msg)
		return m, tea.Batch(cmds...)
	case g.SwitchScreenMsg:
		m.curScreen = g.Screen(msg)
		log.Printf("From systemd, SwitchScreenMsg: %s", m.curScreen.String())
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	}

	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder
	switch m.curScreen {
	case g.Systemd:
		var info string
		if m.allReady {
			info = readyStyle.Render("All units are ready.")
		} else {
			info = notReadyStyle.Render(m.spinner.View())
		}
		fmt.Fprintf(&b, "Watching %d services (%.0fs): %s\n\n",
			m.watcher.NumUnits(),
			m.watcher.Elapsed().Seconds(),
			info,
		)

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

			if i == m.cursorIndex && !m.textinputSelected {
				fmt.Fprintf(&b, "-> ")
			}

			fmt.Fprintf(&b, "%s%s %s\n", left, alignRight(80-len(left), right), u.Description)
		}

		fmt.Fprintf(&b, "\n%s", m.textinput.View())

		return b.String()
	case g.UnitInfoScreen:
		return m.unitInfo.View()
	}
	return "Something went wrong!"
}

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}
