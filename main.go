package main

import (
	"fmt"
	"log"
	"net/http"
	"spirit-box/logging"
	"spirit-box/services"
	"spirit-box/tui"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/dbus"
	"golang.org/x/net/websocket"
)

const PORT = "8080"
const CHANNEL_BUFFER = 100
const SYSTEMD_UPDATE_INTERVAL = 500 // in milliseconds

var web = true

func SocketTest(messages chan interface{}) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		log.Printf("Received socket connection.")
		for {
			msg := <-messages
			websocket.JSON.Send(ws, msg)
		}
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

	if web {
		messages := make(chan interface{}, CHANNEL_BUFFER)
		uw.AttachChannel(messages)

		http.HandleFunc("/socket",
			func(w http.ResponseWriter, req *http.Request) {
				s := websocket.Server{Handler: websocket.Handler(SocketTest(messages))}
				s.ServeHTTP(w, req)
			})

		log.Printf("Server started on port %s.", PORT)

		go uw.Start(SYSTEMD_UPDATE_INTERVAL)

		go func() {
			err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
			if err != nil {
				log.Fatal("ListenAndServe:" + err.Error())
			}
		}()

		time.Sleep(300 * time.Second)
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
