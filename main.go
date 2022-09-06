package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"spirit-box/config"
	"spirit-box/device"
	"spirit-box/logging"
	"spirit-box/scripts"
	"spirit-box/services"
	"spirit-box/tui"
	"spirit-box/tui_lite"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/rs/cors"
)

//go:embed webui/build
var embeddedFiles embed.FS

func init() {
	const (
		defaultPath      = "/etc/spirit-box/"
		pathUsage        = "Path to the directory where spirit-box stores config files and logs."
		defaultDebugFile = "/dev/null"
		debugUsage       = "Write debugging logs to a file."
		defaultTui       = false
		tuiUsage         = "Use fancy TUI. Not recommended for serial consoles."
	)

	flag.StringVar(&config.SPIRIT_PATH, "p", defaultPath, pathUsage)
	flag.StringVar(&config.DEBUG_FILE, "d", defaultDebugFile, debugUsage)
	flag.BoolVar(&config.TUI_FANCY, "tui_fancy", false, tuiUsage)
}

func createSystemdHandler(uw *services.UnitWatcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(uw.Units)
	}
}

func createScriptsHandler(sc *scripts.ScriptController) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sc.PriorityGroups)
	}
}

func createQuitHandler(quit chan struct{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		quit <- struct{}{}
	}
}

// so frontend knows if host machine's default web page is up
func hostUpHandler(w http.ResponseWriter, r *http.Request) {
	if device.HOST_IS_UP {
		fmt.Fprintf(w, "up")
	} else {
		fmt.Fprintf(w, "not up")
	}
}

func getFileSystem() http.FileSystem {
	fsys, err := fs.Sub(embeddedFiles, "webui/build")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}

func main() {
	quitWeb := make(chan struct{})
	quitTui := make(chan struct{})
	rebootServer := make(chan struct{})

	flag.Parse()
	config.InitPaths()

	device.LoadNetworkConfig()

	ip := device.GetIPv4Addr(device.NIC)
	ip = ip[:len(ip)-3]

	// apply iptables rules
	device.UnsetPortForwarding()
	err := device.SetPortForwarding()
	if err != nil {
		log.Fatal(err)
	}

	dConn, err := dbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer dConn.Close()

	logging.InitLogger()
	uw := services.NewWatcher(dConn)
	sc := scripts.NewController()

	// setup endpoints for server
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(getFileSystem()))
	mux.HandleFunc("/systemd", createSystemdHandler(uw))
	mux.HandleFunc("/scripts", createScriptsHandler(sc))
	mux.HandleFunc("/quit", createQuitHandler(quitWeb))
	mux.HandleFunc("/host", hostUpHandler)

	log.Printf("Starting server on port %s.", device.SERVER_PORT)
	handler := cors.Default().Handler(mux)

	// Writes default log messages (log.Print, log.Fatal, etc...)
	// to /dev/null by default. File can be specified with -d flag.
	f, err := tea.LogToFile(config.DEBUG_FILE, "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Printf("\033[2J") // clear the screen
	log.Print("Starting spirit-box...")
	uw.InitializeStates()
	go sc.RunPriorityGroups()

	go func() { // start server, reboot if reboot message is sent
		for {
			s := http.Server{Addr: fmt.Sprintf(":%s", device.SERVER_PORT), Handler: handler}
			go func() {
				time.Sleep(time.Duration(500) * time.Millisecond)
				<-rebootServer
				s.Shutdown(context.Background())
			}()
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("ListenAndServe:" + err.Error())
			}
		}
	}()

	go func() {
		time.Sleep(time.Second)
		for {
			allReady := uw.AllReady()
			if !allReady {
				continue
			}
			allReady = sc.AllReady()

			//res, _ := http.Get(fmt.Sprintf("http://localhost:%s", device.TEMP_PORT))
			if allReady {
				err := device.UnsetPortForwarding()
				if err != nil {
					log.Fatal(err)
				}
				device.HOST_IS_UP = true
				time.Sleep(2 * time.Second)
				rebootServer <- struct{}{}
				break
			}
			time.Sleep(time.Second)
		}
	}()

	var p *tea.Program
	go func(quit chan struct{}) {
		// the tui logic will "pump" the updates of the unit watcher.
		// no need to run uw.Start
		if config.TUI_FANCY {
			p = tui.CreateProgram(dConn, uw, ip, sc)
		} else {
			p = tui_lite.CreateProgram(dConn, uw, ip, sc)
		}
		if err := p.Start(); err != nil {
			fmt.Printf("There was an error: %v\n", err)
			os.Exit(1)
		}
		log.Print("Program exited.")
		quit <- struct{}{}
		log.Print("quit signal sent to channel.")
	}(quitTui)

	select {
	case <-quitWeb:
		p.Quit()
	case <-quitTui:
		break
	}

	log.Print("Cleanup.")
	device.UnsetPortForwarding() // No problems if rules were already unset.
	fmt.Printf("\033[2J")        // clear the screen

	// Dump log lines to stdout for dev purposes.
	fmt.Printf("\nLog Lines (in order of insertion):\n")
	for _, event := range logging.Logs.Events {
		fmt.Println(event.LogLine())
	}

	logFile := logging.CreateLogFile()
	defer logFile.Close()

	logging.Logs.WriteJSON(logFile)
	fmt.Printf("\nWrote JSON log entries to %s.\n", logFile.Name())
}
