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
	substate string // exited or running
	ready bool // observed substate matches desired substate
}

func main() {
	dConn, err := dbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer dConn.Close()

	units := loadWhitelist("/home/severian/data-driven-boot-up-ui/whitelist")
	fmt.Println("Units to be watched:")
	for name, info := range units {
		fmt.Printf("%s, ready when substate=%s\n", name, info.substate)
	}
	//return

	// May be useful for monitoring:
	// SubscribeUnitsCustom 
	// SetPropertiesSubscriber
	// SetSubStateSubscriber
	// SubscriptionSet

	watchAll := false

	timeout := 120
	fmt.Printf("\nTimeout = %ds\n", timeout)
	timer := time.NewTimer(time.Duration(timeout)*time.Second)
	dConn.Subscribe()
	if watchAll {
		watchAllUnits(dConn, timer)
	} else {
		watchUnits(dConn, timer, units)
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

func watchUnits(dConn *dbus.Conn, timer *time.Timer, unitsToWatch map[string]unitInfo) {
	fmt.Println("\nInitial states:")
	subset := dConn.NewSubscriptionSet()
	for unitName, _ := range unitsToWatch {
		properties, err := dConn.GetUnitProperties(unitName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(
			"%s: %s %s %s\n",
			unitName,
			properties["LoadState"],
			properties["ActiveState"],
			properties["SubState"],
		)

		unit := unitsToWatch[unitName]
		if unit.substate == "watch" || properties["SubState"] == unit.substate {
			unit.ready = true
			unitsToWatch[unitName] = unit
		}

		subset.Add(unitName)
	}

	updates, errors := subset.Subscribe()
	for {
		allReady := true
		fmt.Print("\nWaiting for unit updates...\n")
		for unitName, info := range unitsToWatch {
			allReady = allReady && info.ready
			fmt.Printf("%s: ready=%t\n", unitName, info.ready)
		}
		fmt.Println()
		if allReady {
			fmt.Printf("%d units are ready.\n", len(unitsToWatch))
			return
		}
		select {
			case update := <-updates:
				for name, status := range update {
					unit := unitsToWatch[name]
					fmt.Printf("%s was updated: ", name)
					if status == nil {
						fmt.Println("dead")
						if unitsToWatch[name].substate != "dead" {
							break
						}
					} else if unit.substate != "watch" && status.SubState != unit.substate {
						fmt.Printf("%s %s %s\n", status.LoadState, status.ActiveState, status.SubState)
						break
					}
					fmt.Printf("%s %s %s\n", status.LoadState, status.ActiveState, status.SubState)
					unit.ready = true
					unitsToWatch[name] = unit
				}
			case err := <-errors:
				log.Fatal(err)
			case <-timer.C:
				return
		}
	}
}

func loadWhitelist(filename string) map[string]unitInfo {
	units := make(map[string]unitInfo)
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

		units[split[0]] = unitInfo{split[1], false}
	}
	return units
}
