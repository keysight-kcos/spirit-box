package main

import(
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"spirit-box/logging"
	"spirit-box/services"
	"spirit-box/device"
	"log"
	"time"
	// "os/exec"
)

func main() {
	fmt.Printf("\033[2J") // clear the screen

	logFile := logging.InitLogger()
	fmt.Printf("Writing log to %s\n", logFile)

	/*
	cmd := exec.Command("ip", "addr")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(out))
	*/
	device.PrintInterfaces()

	dConn, err := dbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer dConn.Close()

	units := services.LoadWhitelist("/usr/share/spirit-box/whitelist")
	fmt.Println("Units to be watched:")
	for _, info := range units {
		fmt.Printf("%s, ready when substate=%s\n", info.Name, info.SubStateDesired)
	}

	timeout := 30
	fmt.Printf("\nTimeout = %ds\n\n", timeout)
	timer := time.NewTimer(time.Duration(timeout)*time.Second)
	interval := time.Second

	dConn.Subscribe()
	services.WatchUnits(dConn, interval, timer, units)
}
