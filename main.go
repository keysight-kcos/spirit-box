package main

import(
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"log"
	"time"
	"os"
	"bufio"
	"strings"
	"errors"
)

type unitInfo struct {
	name string
	substate string // exited or running
	ready bool // observed substate matches desired substate
}

func main() {
	dConn, err := dbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer dConn.Close()

	units := loadWhitelist("/usr/share/spirit-box/whitelist")
	fmt.Println("Units to be watched:")
	for _, info := range units {
		fmt.Printf("%s, ready when substate=%s\n", info.name, info.substate)
	}

	watchAll := false

	timeout := 30
	fmt.Printf("\nTimeout = %ds\n", timeout)
	timer := time.NewTimer(time.Duration(timeout)*time.Second)
	interval := time.Second

	dConn.Subscribe()
	if watchAll {
		watchAllUnits(dConn, timer)
	} else {
		watchUnits(dConn, interval, timer, units)
	}
}

func watchAllUnits(dConn *dbus.Conn, timer *time.Timer) {
	updates, errors := dConn.SubscribeUnits(time.Second)
	for {
		select {
			case update := <-updates:
				for name, status := range update {
					substate := "stopped"
					if status != nil {
						substate = status.SubState
					}
					fmt.Printf("%s was updated. Substate: %s\n", name, substate)
				}
			case err := <-errors:
				log.Fatal(err)
			case <-timer.C:
				return
		}
	}
}

func watchUnits(
	dConn *dbus.Conn,
	interval time.Duration,
	timer *time.Timer,
	unitsToWatch []*unitInfo,
) {
	fmt.Println("\nInitial states:")
	for _, unitInfo := range unitsToWatch {
		properties, err := dConn.GetUnitProperties(unitInfo.name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(
			"%s: %s %s %s\n",
			unitInfo.name,
			properties["LoadState"],
			properties["ActiveState"],
			properties["SubState"],
		)

		if unitInfo.substate == "watch" || properties["SubState"] == unitInfo.substate {
			unitInfo.ready = true
		}
	}

	started := time.Now()
L:
	for i := 0; ; i++ {
		allReady := true
		eraseToEndOfLine()
		fmt.Printf("\nWaiting for unit updates... (%.4fs)\n", time.Since(started).Seconds())
		for _, unitInfo := range unitsToWatch {
			allReady = allReady && unitInfo.ready
			eraseToEndOfLine()
			fmt.Printf("%s: ready=%t\n", unitInfo.name, unitInfo.ready)
		}
		fmt.Println()
		if allReady {
			if i != 0 {
				moveCursorDown(len(unitsToWatch))
				fmt.Print("\n")
			}
			fmt.Printf("%d units are ready.\n", len(unitsToWatch))
			break
		}
		select {
			case <-timer.C:
				moveCursorDown(len(unitsToWatch))
				fmt.Println("\nTimed out.")
				break L
			default:
				break
		}
		time.Sleep(interval)
		for _, unitInfo := range unitsToWatch {
			properties, err := dConn.GetUnitProperties(unitInfo.name)
			if err != nil {
				log.Fatal(err)
			}
			eraseToEndOfLine()
			fmt.Printf(
				"%s: %s %s %s\n",
				unitInfo.name,
				properties["LoadState"],
				properties["ActiveState"],
				properties["SubState"],
			)

			if unitInfo.substate == "watch" || properties["SubState"] == unitInfo.substate {
				unitInfo.ready = true
			}
		}
		moveCursorUp(2*len(unitsToWatch)+3)
	}
}

func loadWhitelist(filename string) []*unitInfo {
	units := make([]*unitInfo, 0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, ":")
		if len(split) != 2 {
			log.Fatal(errors.New("Line in whitelist did not match <unit name>:<substate> format."))
		}

		units = append(units, &unitInfo{split[0], split[1], false})
	}
	return units
}

func saveCursorPosition() {
	fmt.Print("\033[s")
}

func restoreCursorPosition() {
	fmt.Print("\033[u")
}

func eraseToEndOfLine() {
	fmt.Printf("\033[K")
}

func moveCursorUp(lines int) {
	fmt.Printf("\033[%dA", lines)
}

func moveCursorDown(lines int) {
	fmt.Printf("\033[%dB", lines)
}
