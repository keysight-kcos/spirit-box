package tui_lite

import (
	"fmt"
	"log"
	"strings"
	"time"

	"spirit-box/config"
	"spirit-box/device"
	"spirit-box/scripts"
	"spirit-box/services"
	"spirit-box/styles"
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
	watcher    *services.UnitWatcher
	controller *scripts.ScriptController
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
	case g.WipeScreenMsg:
		log.Print("wipe")
		m.wipe = true
		return m, tea.Batch(cmds...)
	case g.CheckSystemdMsg:
		m.watcher.UpdateAll()
		return m, tea.Batch(cmds...)
	case g.UpdateIPsMsg:
		m.ipStr = device.CreateIPStr()
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
	header, allReady := m.StatusHeader()
	fmt.Fprintf(&b, lp.JoinHorizontal(lp.Top, styles.DoubleBorder.Render("spirit-box"), header))
	fmt.Fprintf(&b, "\n")
	if allReady && config.BANNER_MESSAGE != "" {
		log.Printf(config.BANNER_MESSAGE)
		fmt.Fprintf(&b, lp.PlaceHorizontal(100, 0.0, styles.DoubleBorderPadded.Render(config.BANNER_MESSAGE)))
		fmt.Fprintf(&b, "\n\nPress 'r' to manually re-render the screen.\n")
		return lp.PlaceHorizontal(width, 0, b.String())
	}

	var readyStatus string
	fmt.Fprintf(&b, "\nSystemD Units:\n")
	for _, u := range m.watcher.Units {
		var displayName string
		if u.Desc != "" {
			displayName = u.Desc
		} else {
			displayName = u.Name
		}

		if u.Ready {
			readyStatus = readyStyle.Render("READY")
		} else {
			readyStatus = notReadyStyle.Render(m.spinner.View())
		}

		left := displayName + ":"
		fmt.Fprintf(&b, "%s%s\n", left, alignRight(100-len(left), readyStatus))
	}

	fmt.Fprintf(&b, "\nScripts:\n")
	for _, s := range m.controller.GetScriptStatuses() {
		var displayName string
		if s.Desc != "" {
			displayName = s.Desc
		} else {
			displayName = s.Cmd
		}
		switch s.Status {
		case 0:
			readyStatus = notReadyStyle.Render("NOT STARTED")
		case 2:
			readyStatus = notReadyStyle.Render("FAILED")
		case 3:
			readyStatus = readyStyle.Render("SUCCEEDED")
		default:
			readyStatus = notReadyStyle.Render(m.spinner.View())
		}
		fmt.Fprintf(&b, "%s%s\n", displayName, alignRight(100-len(displayName), readyStatus))
	}

	fmt.Fprintf(&b, "\nPress 'r' to manually re-render the screen.\n")

	return lp.PlaceHorizontal(width, 0, b.String())
}

func (m model) StatusHeader() (string, bool) {
	var b strings.Builder
	var info string
	var allReady bool

	unitsRemaining := m.watcher.NumUnitsNotReady()
	scriptsRemaining, scriptsFailed := m.controller.GetStatus()
	if unitsRemaining == 0 {
		info = readyStyle.Render("All systemd units are ready.")
	} else {
		info = notReadyStyle.Render(fmt.Sprintf("Waiting for %d systemd units to be ready.", unitsRemaining))
	}
	fmt.Fprintf(&b, info)

	if scriptsRemaining == 0 {
		if scriptsFailed == 0 {
			info = readyStyle.Render("\nAll scripts have succeeded.")
		} else {
			info = notReadyStyle.Render(
				fmt.Sprintf("\n%d out of %d scripts have failed.", scriptsFailed, m.controller.NumScripts))
		}
	} else {
		info = notReadyStyle.Render(
			fmt.Sprintf("\nWaiting for %d scripts to finish. (%d failed)", scriptsRemaining, scriptsFailed))
	}
	fmt.Fprintf(&b, info)

	fmt.Fprintf(&b, "\n")
	if unitsRemaining == 0 && scriptsRemaining == 0 && scriptsFailed == 0 {
		fmt.Fprintf(&b, styles.Blinking.Render("System is ready. Press 'q' to close spirit-box."))
		allReady = true
	}

	fmt.Fprintf(&b, fmt.Sprintf("\n%s\n", m.ipStr))

	return styles.LeftPadding.Render(b.String()), allReady
}

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}

func initialModel(dConn *dbus.Conn, watcher *services.UnitWatcher, sc *scripts.ScriptController) model {
	s := spinner.New()
	s.Spinner = spinner.Line
	whitespace := ""
	for i := 0; i < height; i++ {
		whitespace += "\n"
	}
	return model{
		watcher:    watcher,
		controller: sc,
		ipStr:      device.CreateIPStr(),
		spinner:    s,
		whitespace: whitespace,
	}
}

func CreateProgram(dConn *dbus.Conn, watcher *services.UnitWatcher, sc *scripts.ScriptController) *tea.Program {
	model := initialModel(dConn, watcher, sc)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// update ticker
	go func(p *tea.Program) {
		for {
			p.Send(g.CheckSystemdMsg(struct{}{}))
			time.Sleep(time.Second)
		}
	}(p)
	go func(p *tea.Program) {
		for {
			p.Send(g.UpdateIPsMsg(struct{}{}))
			time.Sleep(3*time.Second)
		}
	}(p)
	return p
}
