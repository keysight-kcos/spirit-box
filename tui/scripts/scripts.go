package scripts

import (
	"fmt"
	"log"
	g "spirit-box/tui/globals"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	cursorIndex int
	choices []string
	selected map[int]struct{}
	curScreen g.Screen
}

func New() Model {
	return Model{
		choices: []string{"script1", "script2", "script3"},
		selected: make(map[int]struct{}),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursorIndex < len(m.choices)-1 {
				m.cursorIndex++
			}
		case "k", "up":
			if m.cursorIndex > 0 {
				m.cursorIndex--
			}
		case "enter":
			_, ok := m.selected[m.cursorIndex]
			if ok {
				delete(m.selected, m.cursorIndex)
			} else {
				m.selected[m.cursorIndex] = struct{}{}
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
    s := "View which script?\n\n"

    for i, choice := range m.choices {

        cursor := " "
        if m.cursorIndex == i {
            cursor = ">"
        }

        checked := " "
        if _, ok := m.selected[i]; ok {
            checked = "x"
        }

        s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
    }

    s += "\nPress q to return.\n"

    return s
}
