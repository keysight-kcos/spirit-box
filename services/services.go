// For the observation of systemd services.
package services

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"log"
	"os"
	"spirit-box/logging"
	"strings"
	"time"
)

type UnitInfo struct {
	Name            string
	SubStateDesired string // service will be considered ready when this substate is met. 
						   // set to "any" if any substate is okay.
	Ready           bool   // observed substate matches desired substate
	LoadState       string
	ActiveState     string
	SubState string
}

func (u *UnitInfo) printAndLogUnitUpdates(updates [3]string, l *log.Logger) {
	changed := false
	if updates[0] != u.LoadState {
		changed = true
		u.LoadState = updates[0]
	}
	if updates[1] != u.ActiveState {
		changed = true
		u.ActiveState = updates[1]
	}
	if updates[2] != u.SubState {
		changed = true
		u.SubState = updates[2]
	}

	statline := fmt.Sprintf(
		"%s: %s %s %s",
		u.Name,
		u.LoadState,
		u.ActiveState,
		u.SubState,
	)
	eraseToEndOfLine()
	fmt.Print(statline+"\n")

	if u.SubStateDesired == "any" || u.SubState == u.SubStateDesired {
		u.Ready = true
		statline += " READY"
	}

	if changed {
		l.Print(statline+"\n")
	}

}

func printAndLogUnits(dConn *dbus.Conn, l *log.Logger, units []*UnitInfo) {
	for _, u := range units {
		properties, err := dConn.GetUnitProperties(u.Name)
		if err != nil {
			log.Fatal(err)
		}

		// type assertions
		s1, ok := properties["LoadState"].(string)
		if !ok {
			log.Fatal(errors.New("Type assertion failed: properties[\"LoadState\"] is not a string."))
		}
		s2, ok := properties["ActiveState"].(string)
		if !ok {
			log.Fatal(errors.New("Type assertion failed: properties[\"ActiveState\"] is not a string."))
		}
		s3, ok := properties["SubState"].(string)
		if !ok {
			log.Fatal(errors.New("Type assertion failed: properties[\"SubState\"] is not a string."))
		}

		u.printAndLogUnitUpdates([3]string{s1, s2, s3}, l)
	}
}

func WatchUnits(
	dConn *dbus.Conn,
	interval time.Duration,
	timer *time.Timer,
	unitsToWatch []*UnitInfo,
) {
	l := logging.Logger

	started := time.Now()
	fmt.Println("Current unit statuses:")

L:
	for i := 0; ; i++ {
		printAndLogUnits(dConn, l, unitsToWatch)
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
				//moveCursorDown(len(unitsToWatch))
				fmt.Print("\n")
			}
			fmt.Printf("%d units are ready.\n", len(unitsToWatch))
			break
		}
		select {
		case <-timer.C:
			//moveCursorDown(len(unitsToWatch))
			fmt.Println("\nTimed out.\n")
			break L
		default:
			break
		}
		time.Sleep(interval)
		//printAndLogUnits(dConn, l, unitsToWatch)
		moveCursorUp(2*len(unitsToWatch) + 3)
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

		units = append(units, &UnitInfo{split[0], split[1], false, "", "", ""})
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
