// For the observation of systemd services.
package services

import (
	"bufio"
	"errors"
	"github.com/coreos/go-systemd/v22/dbus"
	"log"
	"os"
	"strings"
	"time"
)

const whitelistPath = "/usr/share/spirit-box/whitelist"

type UnitWatcher struct {
	Units []*UnitInfo
	dConn *dbus.Conn
	started time.Time
}

func (uw *UnitWatcher) UpdateAll() {
	for _, u := range uw.Units {
		properties, err := uw.dConn.GetUnitProperties(u.Name)
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

		u.update([3]string{s1, s2, s3})
	}
}

func (uw *UnitWatcher) Elapsed() time.Duration {
	return time.Since(uw.started)
}

func (uw *UnitWatcher) NumUnits() int {
	return len(uw.Units)
}

func NewWatcher(dConn *dbus.Conn) *UnitWatcher {
	units := LoadWhitelist(whitelistPath)
	return &UnitWatcher{
		dConn: dConn,
		Units: units,
		started: time.Now(),
	}
}

type UnitInfo struct {
	Name            string
	SubStateDesired string // service will be considered ready when this substate is met. 
						   // set to "any" if any substate is okay.
	Ready           bool   // observed substate matches desired substate
	LoadState       string
	ActiveState     string
	SubState string
}

// Check if unit info needs to be updated, report if it was changed.
func (u *UnitInfo) update(updates[3]string) bool {
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

	if u.SubStateDesired == "any" || u.SubState == u.SubStateDesired {
		u.Ready = true
	}

	return changed
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
