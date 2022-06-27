// For the observation of systemd services.
package services

import (
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"time"
	"log"
	"os"
	"bufio"
	"strings"
	"errors"
)

type UnitInfo struct {
	Name string
	Substate string // exited or running
	Ready bool // observed substate matches desired substate
}

func WatchUnits(
	dConn *dbus.Conn,
	interval time.Duration,
	timer *time.Timer,
	unitsToWatch []*UnitInfo,
) {
	fmt.Println("\nInitial states:")
	for _, UnitInfo := range unitsToWatch {
		properties, err := dConn.GetUnitProperties(UnitInfo.Name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Sprintf(
			"%s: %s %s %s\n",
			UnitInfo.Name,
			properties["LoadState"],
			properties["ActiveState"],
			properties["SubState"],
		)

		if UnitInfo.Substate == "watch" || properties["SubState"] == UnitInfo.Substate {
			UnitInfo.Ready = true
		}
	}

	started := time.Now()
L:
	for i := 0; ; i++ {
		allReady := true
		eraseToEndOfLine()
		fmt.Printf("\nWaiting for unit updates... (%.4fs)\n", time.Since(started).Seconds())
		for _, UnitInfo := range unitsToWatch {
			allReady = allReady && UnitInfo.Ready
			eraseToEndOfLine()
			fmt.Printf("%s: ready=%t\n", UnitInfo.Name, UnitInfo.Ready)
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
		for _, UnitInfo := range unitsToWatch {
			properties, err := dConn.GetUnitProperties(UnitInfo.Name)
			if err != nil {
				log.Fatal(err)
			}
			eraseToEndOfLine()
			fmt.Printf(
				"%s: %s %s %s\n",
				UnitInfo.Name,
				properties["LoadState"],
				properties["ActiveState"],
				properties["SubState"],
			)

			if UnitInfo.Substate == "watch" || properties["SubState"] == UnitInfo.Substate {
				UnitInfo.Ready = true
			}
		}
		moveCursorUp(2*len(unitsToWatch)+3)
	}
}

func LoadWhitelist(filename string) []*UnitInfo {
	units := make([]*UnitInfo, 0)
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

		units = append(units, &UnitInfo{split[0], split[1], false})
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
