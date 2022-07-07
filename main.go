package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"spirit-box/logging"
	"spirit-box/services"
	"spirit-box/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/rs/cors"
)

const PORT = "8080"
const CHANNEL_BUFFER = 100
const SYSTEMD_UPDATE_INTERVAL = 500 // in milliseconds

var web = true

func createSystemdHandler(uw *services.UnitWatcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(uw.Units)
	}
}

func createQuitHandler(quit chan struct{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		quit <- struct{}{}
	}
}

func main() {
	dConn, err := dbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer dConn.Close()

	logging.InitLogger()

	// create a new watcher, add channels that will receive updates, start it up with an update interval
	uw := services.NewWatcher(dConn)

	quit := make(chan struct{})
	if web {
		go uw.Start(SYSTEMD_UPDATE_INTERVAL)

		mux := http.NewServeMux()
		mux.HandleFunc("/systemd", createSystemdHandler(uw))
		mux.HandleFunc("/quit", createQuitHandler(quit))

		log.Printf("Starting server on port %s.", PORT)
		handler := cors.Default().Handler(mux)

		go func() {
			err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), handler)
			if err != nil {
				log.Fatal("ListenAndServe:" + err.Error())
			}
		}()
	} else {
		fmt.Printf("\033[2J") // clear the screen

		// Writes default log messages (log.Print, log.Fatal, etc...)
		// to a file called tuiDebug.
		f, err := tea.LogToFile("tuiDebug", "debug")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.Print("Starting spirit-box...")

		uw.InitializeStates()
		// the tui logic will "pump" the updates of the unit watcher.
		// no need to run uw.Start
		tui.StartTUI(dConn, uw)
	}

	<-quit

	// Dump log lines to stdout for dev purposes.
	fmt.Printf("\nLog Lines (%d):\n", logging.Logs.Length())
	for _, event := range logging.Logs.Events {
		fmt.Println(event.LogLine())
	}

	logFile := logging.CreateLogFile()
	defer logFile.Close()

	logging.Logs.WriteJSON(logFile)
	fmt.Printf("\nWrote JSON log entries to %s.\n", logFile.Name())
}
