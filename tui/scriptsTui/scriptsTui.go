package scriptsTui

import (
	"fmt"
	"log"
	"strings"
	"spirit-box/scripts"
	g "spirit-box/tui/globals"
	tea "github.com/charmbracelet/bubbletea"
	lp "github.com/charmbracelet/lipgloss"
)

var readyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("10"))
var notReadyStyle = lp.NewStyle().Bold(true).Foreground(lp.Color("9"))
var alignRightStyle = lp.NewStyle().Align(lp.Right)
var alignLeftStyle = lp.NewStyle().Align(lp.Left)

type Model struct {
	cursorIndex int
	curScreen g.Screen
	sc *scripts.ScriptController
	AllReady bool
}

func New(sc *scripts.ScriptController) Model {
	return Model{
		sc: sc,
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
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			cmd := func() tea.Msg { return g.SwitchScreenMsg(g.TopLevel) }
			cmds = append(cmds, cmd)
		}
	}

	switch msg := msg.(type) {
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
		info = notReadyStyle.Render("Scripts not ready.")
	}
	fmt.Fprintf(&b, "Watching %d priority groups: %s\n\n",
		len(m.sc.PriorityGroups),
		info,
	)

	var readyStatus string
	for i, u := range m.sc.PriorityGroups {
		readyStatus = readyStyle.Render("WATCHING")
		left := fmt.Sprintf("PG %d:", u.Num)
		right := fmt.Sprintf("%s",
			alignRight(len("WATCHING"), readyStatus),
		)

		if i == m.cursorIndex {
			fmt.Fprintf(&b, "-> ")
		}

		fmt.Fprintf(&b, "%s%s\n", left, alignRight(20-len(left), right))
	}

	return b.String()
}

func alignRight(width int, str string) string {
	return alignRightStyle.Width(width).Render(str)
}

