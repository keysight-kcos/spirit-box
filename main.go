package main

import (
	"fmt"
	"log"
	"spirit-box/logging"
	"spirit-box/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/dbus"
)

func main() {
	fmt.Printf("\033[2J") // clear the screen

	// Writes default log messages (log.Print, log.Fatal, etc...)
	// to a file called tuiDebug.
	f, err := tea.LogToFile("tuiDebug", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.Print("Starting spirit-box...")

	logging.InitLogger()
	logFile := logging.CreateLogFile()
	defer logFile.Close()

	dConn, err := dbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer dConn.Close()

	tui.StartTUI(dConn)

	fmt.Printf("\nWrote JSON log entries to %s.\n", logFile.Name())

	// Dump log lines to stdout for dev purposes.
	fmt.Printf("\nLog Lines (%d):\n", logging.Logs.Length())
	for _, event := range logging.Logs.Events {
		fmt.Println(event.LogLine())
	}

	logging.Logs.WriteJSON(logFile)
}
