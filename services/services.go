// For the observation of systemd services.
package services

import (
	"bufio"
	"errors"
	"github.com/coreos/go-systemd/v22/dbus"
	"log"
	"spirit-box/logging"
	"fmt"
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

func (uw *UnitWatcher) InitializeStates() {
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
		s4, ok := properties["Description"].(string)
		if !ok {
			log.Fatal(errors.New("Type assertion failed: properties[\"Description\"] is not a string."))
		}
		u.Description = s4

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
	newUW := &UnitWatcher{
		dConn: dConn,
		Units: units,
		started: time.Now(),
	}
	newUW.InitializeStates()
	return newUW
}

type UnitInfo struct {
	Name            string
	SubStateDesired string // service will be considered ready when this substate is met. 
						   // set to "any" if any substate is okay.
	Ready           bool   // observed substate matches desired substate
	LoadState       string
	ActiveState     string
	SubState        string
	Description     string
	At              time.Time
}

// Check if unit info needs to be updated, log if it was changed.
func (u *UnitInfo) update(updates[3]string) {
	from1, from2, from3, from4 := u.LoadState, u.ActiveState, u.SubState, u.Ready
	changed := false
	if updates[0] != u.LoadState {
		u.LoadState = updates[0]
		changed = true
	}
	if updates[1] != u.ActiveState {
		u.ActiveState = updates[1]
		changed = true
	}
	if updates[2] != u.SubState {
		u.SubState = updates[2]
		changed = true
	}

	if u.SubStateDesired == "any" || u.SubState == u.SubStateDesired {
		u.Ready = true
	} else {
		u.Ready = false
	}

	if changed {
		obj := u.GetStateChange(from1, from2, from3, from4)
		le := logging.NewLogEvent(fmt.Sprintf("%s state change.", u.Name), obj)
		le.EndTime = time.Now()
		le.StartTime = u.At
		u.At = le.EndTime
		logging.Logs.AddLogEvent(le)
	}
}

type UnitStateChange struct {
	Name            string
	SubStateDesired string
	Ready           [2]bool
	LoadState       [2]string
	ActiveState     [2]string
	SubState        [2]string
	Description     string
}

func (u *UnitInfo) GetStateChange(from1, from2, from3 string, from4 bool) *UnitStateChange {
	return &UnitStateChange{
		Name: u.Name,
		SubStateDesired: u.SubStateDesired,
		LoadState: [2]string{from1, u.LoadState},
		ActiveState: [2]string{from2, u.ActiveState},
		SubState: [2]string{from3, u.SubState},
		Ready: [2]bool{from4, u.Ready},
		Description: u.Description,
	}
}

func (u *UnitStateChange) LogLine() string {
	return fmt.Sprintf("%s: %s %s %s %s", u.Name, u.LoadState[1], u.ActiveState[1], u.SubState[1], u.Description)
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

		units = append(units, &UnitInfo{split[0], split[1], false, "", "", "", "", time.Now()})
	}
	return units
}
