package scriptsTui

import (
	"fmt"
	"log"
	"spirit-box/scripts"
	g "spirit-box/tui/globals"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	lp "github.com/charmbracelet/lipgloss"
)

var readyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("10"))
var notReadyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("9"))
var alignRightStyle = lp.NewStyle().Align(lp.Right)
var alignLeftStyle = lp.NewStyle().Align(lp.Left)

type Model struct {
	cursorIndex   int
	curScreen     g.Screen
	sc            *scripts.ScriptController
	AllReady      bool
	openPgs       []bool
	scriptCursors []int
}

func New(sc *scripts.ScriptController) Model {
	openPgs := make([]bool, len(sc.PriorityGroups))
	scriptCursors := make([]int, len(sc.PriorityGroups))
	/*
		for i, _ := range temp {
			temp[i] = true
		}
	*/
	return Model{
		sc:            sc,
		openPgs:       openPgs,
		scriptCursors: scriptCursors,
		AllReady:      false,
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
			return m, nil
		case "k", "up":
			if m.cursorIndex > 0 {
				m.cursorIndex--
			}
			return m, nil
		case "right":
			if m.openPgs[m.cursorIndex] {
				if m.scriptCursors[m.cursorIndex] < len(m.sc.PriorityGroups[m.cursorIndex].Specs)-1 {
					m.scriptCursors[m.cursorIndex]++
				}
			}
			return m, nil
		case "left":
			if m.openPgs[m.cursorIndex] {
				if m.scriptCursors[m.cursorIndex] > 0 {
					m.scriptCursors[m.cursorIndex]--
				}
			}
			return m, nil
		case "enter":
			m.openPgs[m.cursorIndex] = !m.openPgs[m.cursorIndex]
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			return m, func() tea.Msg { return g.SwitchScreenMsg(g.TopLevel) }
		}
	}

	switch msg := msg.(type) {
	case g.CheckScriptsMsg:
		for i, pg := range m.sc.PriorityGroups {
			running, numFailed := pg.GetStatus()
			if running > 0 || numFailed > 0 {
				m.AllReady = false
				break
			}
			if i == len(m.sc.PriorityGroups)-1 {
				m.AllReady = true
			}
		}
		log.Printf("Scripts AllReady: %t", m.AllReady)

		return m, nil
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
			left := fmt.Sprintf("Command")
			right := fmt.Sprintf("%s  %s  %s  %s",
				alignRight(len("# Runs"), "# Runs"),
				alignRight(len("Retry Timeout"), "Retry Timeout"),
				alignRight(len("Total Timeout"), "Total Timeout"),
				alignRight(len("Result"), "Result"),
			)
			fmt.Fprintf(&b, "\t  %s %s\n", alignLeft(longestCmd, left), alignRight(len(right)+2, right))
			for j, spec := range pg.Specs {
				cmdStr := spec.ToString()
				if j == m.scriptCursors[m.cursorIndex] && (m.cursorIndex == i) {
					cmdStr = "-> " + cmdStr
				}
				numRuns := 0
				readyStatus = notReadyStyle.Render("Awaiting execution.")
				if pg.Trackers != nil {
					tracker := pg.Trackers[j]
					numRuns = len(tracker.Runs)
					if !tracker.Finished {
						readyStatus = notReadyStyle.Render("Running...")
					} else if tracker.Succeeded() {
						readyStatus = readyStyle.Render("Succeeded")
					} else {
						readyStatus = notReadyStyle.Render("Failed   ")
					}
				}
				right := fmt.Sprintf("%s   %s   %s   %s",
					alignRight(len("# Runs"), fmt.Sprintf("%d", numRuns)),
					alignRight(len("Retry Timeout"), fmt.Sprintf("%d", spec.RetryTimeout)),
					alignRight(len("Total Timeout"), fmt.Sprintf("%d", spec.TotalWaitTime)),
					alignRight(0, readyStatus),
				)
				fmt.Fprintf(&b, "\t  %s %s\n", alignLeft(longestCmd+len("-> "), cmdStr), right)
			}
			fmt.Fprintf(&b, "\n")
		}
	}

	/*
		for i := 0; i < 20; i++ {
			fmt.Fprintf(&b, "%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%\n")
		}
	*/

	return b.String()
}

type ScriptStatus struct {
	Cmd    string
	Status int // 0: waiting 1: running 2: failed, 3: succeeded
}

// just get statuses of individual scripts for displaying in the top level.
func (m Model) GetScriptStatuses() []ScriptStatus {
	ret := make([]ScriptStatus, m.sc.NumScripts())

	for _, pg := range m.sc.PriorityGroups {
		for j, spec := range pg.Specs {
			cmdStr := spec.ToString()
			stat := 0

			if pg.Trackers != nil {
				tracker := pg.Trackers[j]
				if tracker.Finished {
					if tracker.Succeeded() {
						stat = 3
					} else {
						stat = 2
					}
				} else {
					stat = 1
				}
			}

			ret = append(ret, ScriptStatus{Cmd: cmdStr, Status: stat})
		}
	}

	return ret
}

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}

func alignLeft(width int, str string) string {
	return alignLeftStyle.Width(width).Render(str)
}
