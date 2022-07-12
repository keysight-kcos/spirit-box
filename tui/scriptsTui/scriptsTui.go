package scriptsTui

import (
	"fmt"
	"log"
	"spirit-box/scripts"
	g "spirit-box/tui/globals"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lp "github.com/charmbracelet/lipgloss"
)

var readyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("10"))
var notReadyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("9"))
var alignRightStyle = lp.NewStyle().Align(lp.Right)
var alignLeftStyle = lp.NewStyle().Align(lp.Left)

const UPDATE_INTERVAL = 500 // update interval in milliseconds

type Model struct {
	cursorIndex int
	curScreen   g.Screen
	sc          *scripts.ScriptController
	AllReady    bool
	openPgs     []bool
}

func New(sc *scripts.ScriptController) Model {
	temp := make([]bool, len(sc.PriorityGroups))
	for i, _ := range temp {
		temp[i] = true
	}
	return Model{
		sc:       sc,
		openPgs:  temp,
		AllReady: false,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursorIndex < len(m.sc.PriorityGroups)-1 {
				m.cursorIndex++
			}
		case "k", "up":
			if m.cursorIndex > 0 {
				m.cursorIndex--
			}
		case "enter":
			m.openPgs[m.cursorIndex] = !m.openPgs[m.cursorIndex]
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			cmd := func() tea.Msg { return g.SwitchScreenMsg(g.TopLevel) }
			cmds = append(cmds, cmd)
		}
	}

	switch msg := msg.(type) {
	case g.CheckScriptsMsg:
		m.AllReady = true
		for _, pg := range m.sc.PriorityGroups {
			m.AllReady = m.AllReady && pg.AllSucceeded()
			// add something for counting how many failed
		}

		cmd := func() tea.Msg {
			time.Sleep(time.Duration(UPDATE_INTERVAL) * time.Millisecond)
			return g.CheckScriptsMsg(struct{}{})
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case g.SwitchScreenMsg:
		m.curScreen = g.Screen(msg)
		log.Printf("From scripts, SwitchScreenMsg: %s", m.curScreen.String())
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder
	var info string
	if m.AllReady {
		info = readyStyle.Render("All scripts are ready.")
	} else {
		info = notReadyStyle.Render("Scripts are not ready.")
	}
	fmt.Fprintf(&b, "Watching %d priority groups: %s\n\n",
		len(m.sc.PriorityGroups),
		info,
	)

	var readyStatus string
	longestCmd := m.sc.GetLongestCmdLength()
	for i, pg := range m.sc.PriorityGroups {
		if pg.Trackers == nil {
			readyStatus = notReadyStyle.Render("Awaiting execution")
		} else {
			running, numFailed := pg.GetStatus()
			if running == 0 && numFailed == 0 {
				readyStatus = readyStyle.Render("All scripts succeeded.")
			} else if running == 0 {
				readyStatus = notReadyStyle.Render(fmt.Sprintf("%d scripts failed.", numFailed))
			} else {
				readyStatus = notReadyStyle.Render(
					fmt.Sprintf("Waiting for %d scripts to finish. (%d failed)", running, numFailed))
			}
		}

		left := fmt.Sprintf("Priority Group #%d:", pg.Num)
		right := fmt.Sprintf(" %s", readyStatus)

		if i == m.cursorIndex {
			left = "-> " + left
		}
		fmt.Fprintf(&b, "     %s%s\n", left, right)

		if m.openPgs[i] {
			fmt.Fprintf(&b, "\n")
			for j, spec := range pg.Specs {
				cmdStr := spec.ToString()
				readyStatus = notReadyStyle.Render("Awaiting execution.")
				if pg.Trackers != nil {
					tracker := pg.Trackers[j]
					if !tracker.Finished {
						readyStatus = notReadyStyle.Render("Running...")
					} else if tracker.Succeeded() {
						readyStatus = readyStyle.Render("Succeeded")
					} else {
						readyStatus = notReadyStyle.Render("Failed")
					}
				}
				fmt.Fprintf(&b, "        %s %s\n", alignLeft(longestCmd, cmdStr), alignRight(20, readyStatus))
			}
			fmt.Fprintf(&b, "\n")
		}
	}

	return b.String()
}

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}

func alignLeft(width int, str string) string {
	return alignLeftStyle.Width(width).Render(str)
}
