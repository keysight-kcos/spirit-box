// For the observation of systemd services.
package services

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"spirit-box/logging"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
)

const whitelistPath = "/usr/share/spirit-box/whitelist"

type UnitWatcher struct {
	Units          []*UnitInfo
	DConn          *dbus.Conn
	updateChannels []chan interface{}
	started        time.Time
	mu             sync.Mutex
}

func (uw *UnitWatcher) Start(interval int) {
	time.Sleep(time.Second)
	uw.InitializeStates()
	for {
		time.Sleep(time.Duration(interval) * time.Millisecond)
		uw.UpdateAll()
	}
}

func (uw *UnitWatcher) UpdateAll() bool {
	uw.mu.Lock()
	defer uw.mu.Unlock()
	allReady := true
	for _, u := range uw.Units {
		properties, err := uw.DConn.GetUnitProperties(u.Name)
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

		u.update([3]string{s1, s2, s3}, properties)
		allReady = allReady && u.Ready
	}

	return allReady
}

func (uw *UnitWatcher) InitializeStates() bool {
	uw.mu.Lock()
	defer uw.mu.Unlock()
	allReady := true
	for _, u := range uw.Units {
		uw.InitializeState(u)
		allReady = allReady && u.Ready
	}
	return allReady
}

func (uw *UnitWatcher) InitializeState(u *UnitInfo) {
	properties, err := uw.DConn.GetUnitProperties(u.Name)
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

	u.update([3]string{s1, s2, s3}, properties)
}

func (uw *UnitWatcher) AddUnit(name string) {
	uw.mu.Lock()
	defer uw.mu.Unlock()
	newUnit := &UnitInfo{name, "watch", false, "", "", "", "", nil, time.Now(), uw}
	uw.InitializeState(newUnit)
	uw.Units = append(uw.Units, newUnit)
	// Change logic for initialization?
}

func (uw *UnitWatcher) Elapsed() time.Duration {
	return time.Since(uw.started)
}

func (uw *UnitWatcher) NumUnits() int {
	return len(uw.Units)
}

func (uw *UnitWatcher) AttachChannel(c chan interface{}) {
	uw.mu.Lock()
	defer uw.mu.Unlock()
	uw.updateChannels = append(uw.updateChannels, c)
}

func NewWatcher(dConn *dbus.Conn) *UnitWatcher {
	newUW := &UnitWatcher{
		DConn:          dConn,
		updateChannels: make([]chan interface{}, 0, 5),
		started:        time.Now(),
	}
	newUW.Units = LoadWhitelist(whitelistPath, newUW)

	return newUW
}

// Basic data for a unit's state.
type UnitInfo struct {
	Name            string
	SubStateDesired string // service will be considered ready when this substate is met.
	// set to "any" if any substate is okay.
	Ready       bool // observed substate matches desired substate
	LoadState   string
	ActiveState string
	SubState    string
	Description string
	Properties  map[string]interface{}
	At          time.Time
	uw          *UnitWatcher
}

// Check if unit info needs to be updated, log if it was changed.
func (u *UnitInfo) update(updates [3]string, properties map[string]interface{}) bool {
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

	if u.SubStateDesired == "watch" || u.SubState == u.SubStateDesired {
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

		u.Properties = properties

		for _, c := range u.uw.updateChannels {
			c <- *u
		}
	}

	return changed
}

type UnitStateChange struct {
	Name            string    `json:"name"`
	SubStateDesired string    `json:"subStateDesired"`
	Ready           [2]bool   `json:"ready"`
	LoadState       [2]string `json:"loadState"`
	ActiveState     [2]string `json:"activeState"`
	SubState        [2]string `json:"subState"`
	Description     string    `json:"description"`
}

func (u *UnitInfo) GetStateChange(from1, from2, from3 string, from4 bool) *UnitStateChange {
	return &UnitStateChange{
		Name:            u.Name,
		SubStateDesired: u.SubStateDesired,
		LoadState:       [2]string{from1, u.LoadState},
		ActiveState:     [2]string{from2, u.ActiveState},
		SubState:        [2]string{from3, u.SubState},
		Ready:           [2]bool{from4, u.Ready},
		Description:     u.Description,
	}
}

func (u *UnitStateChange) LogLine() string {
	return fmt.Sprintf("%s: %s %s %s %s", u.Name, u.LoadState[1], u.ActiveState[1], u.SubState[1], u.Description)
}

func (u *UnitStateChange) GetObjType() string {
	return "SystemD Unit state change."
}

func LoadWhitelist(filename string, uw *UnitWatcher) []*UnitInfo {
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

		units = append(units, &UnitInfo{split[0], split[1], false, "", "", "", "", nil, time.Now(), uw})
	}
	return units
}
