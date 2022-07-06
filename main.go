package main

import (
	"fmt"
	"log"
	"net/http"
	"spirit-box/services"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"golang.org/x/net/websocket"
)

const PORT = "8080"
const CHANNEL_BUFFER = 100

var webOnly = true

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

	if webOnly {
		uw := services.NewWatcher(dConn)
		messages := make(chan interface{}, CHANNEL_BUFFER)
		uw.AttachChannel(messages)

		http.HandleFunc("/socket",
			func(w http.ResponseWriter, req *http.Request) {
				s := websocket.Server{Handler: websocket.Handler(SocketTest(messages))}
				s.ServeHTTP(w, req)
			})

		log.Printf("Server started on port %s.", PORT)

		go func() {
			time.Sleep(time.Second)
			uw.InitializeStates()
			for {
				time.Sleep(time.Second)
				uw.UpdateAll()
			}
		}()

		/*
			go func() {
				for {
					time.Sleep(time.Second)
					messages <- []byte("is this thing on?")
				}
			}()
		*/
		err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
		if err != nil {
			log.Fatal("ListenAndServe:" + err.Error())
		}
	} else {
		fmt.Printf("\033[2J") // clear the screen

		/*
			// Writes default log messages (log.Print, log.Fatal, etc...)
			// to a file called tuiDebug.
			f, err := tea.LogToFile("tuiDebug", "debug")
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			log.Print("Starting spirit-box...")

			logging.InitLogger()

			dConn, err := dbus.New()
			if err != nil {
				log.Fatal(err)
			}
			defer dConn.Close()

			tui.StartTUI(dConn)

			// Dump log lines to stdout for dev purposes.
			fmt.Printf("\nLog Lines (%d):\n", logging.Logs.Length())
			for _, event := range logging.Logs.Events {
				fmt.Println(event.LogLine())
			}

			logFile := logging.CreateLogFile()
			defer logFile.Close()

			logging.Logs.WriteJSON(logFile)
			fmt.Printf("\nWrote JSON log entries to %s.\n", logFile.Name())
		*/
	}
}
