package main

import (
	"log"
	"github.com/coreos/go-systemd/v22/dbus"
	tea "github.com/charmbracelet/bubbletea"
	"spirit-box/tui"
)

func main() {
	f, err := tea.LogToFile("tuiDebug", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.Print("starting")

	dConn, err := dbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer dConn.Close()

	tui.StartTUI(dConn)
}
