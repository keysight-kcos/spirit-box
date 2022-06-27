package main

import(
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"spirit-box/logging"
	"spirit-box/services"
	"log"
	"time"
)

func main() {
	logging.InitLogger()
	fmt.Printf("\033[2J") // clear the screen

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
	fmt.Printf("\nTimeout = %ds\n", timeout)
	timer := time.NewTimer(time.Duration(timeout)*time.Second)
	interval := time.Second

	dConn.Subscribe()
	services.WatchUnits(dConn, interval, timer, units)
}
