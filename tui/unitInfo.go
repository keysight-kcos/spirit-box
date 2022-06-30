// model for the UnitInfo page. More information for a specified unit.
package tui

import (
	"fmt"
	"log"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/dbus"
)

type selectedUnit struct {
	name       string
	keys       []string
	properties map[string]interface{}
}

func InitSelectedUnit(dConn *dbus.Conn, name string) selectedUnit {
	u := selectedUnit{}
	var err error
	u.name = name
	u.properties, err = dConn.GetUnitProperties(u.name)
	if err != nil {
		log.Fatal(err)
	}

	u.keys = make([]string, len(u.properties))
	i := 0
	for k := range u.properties {
		u.keys[i] = k
		i++
	}
	sort.Slice(u.keys, func(i, j int) bool {
		return u.keys[i] < u.keys[j]
	})

	return u
}

func (u selectedUnit) Update(msg tea.Msg) (selectedUnit, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return u, func() tea.Msg { return switchScreenMsg(Services) }
		}
	}
	return u, nil
}

func (u selectedUnit) View() string {
	var b strings.Builder
	for _, key := range u.keys {
		v := u.properties[key]
		fmt.Fprintf(&b, "%s: %v\n", key, v)
	}
	return b.String()
}
