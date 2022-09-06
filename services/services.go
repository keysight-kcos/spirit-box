// For the observation of systemd services.
package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"spirit-box/config"
	"spirit-box/logging"
	"sync"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
)

var SYSTEMD_START_TIME time.Time

type UnitSpec struct {
	Name            string `json:"name"`
	Desc            string `json:"desc"`
	SubStateDesired string `json:"subStateDesired"`
}

type UnitWatcher struct {
	Units   []*UnitInfo
	DConn   *dbus.Conn
	started time.Time
	mu      sync.Mutex
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
		s1 := assertString(properties["LoadState"])
		s2 := assertString(properties["ActiveState"])
		s3 := assertString(properties["SubState"])

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

func (uw *UnitWatcher) InitializeState(u *UnitInfo) error {
	properties, err := uw.DConn.GetUnitProperties(u.Name)
	if err != nil {
		return err
	}

	// type assertions
	s1 := assertString(properties["LoadState"])
	s2 := assertString(properties["ActiveState"])
	s3 := assertString(properties["SubState"])
	s4 := assertString(properties["Description"])
	u.Description = s4

	u.update([3]string{s1, s2, s3}, properties)
	return nil
}

func (uw *UnitWatcher) AddUnit(name string) {
	uw.mu.Lock()
	defer uw.mu.Unlock()
	newUnit := &UnitInfo{name, "watch", false, "", "", "", "", "", nil, SYSTEMD_START_TIME, uw}
	err := uw.InitializeState(newUnit)
	if err != nil {
		return // no feedback on failure
	}
	uw.Units = append(uw.Units, newUnit)
}

func (uw *UnitWatcher) AllReadyStatus() string {
	uw.mu.Lock()
	units := uw.Units
	uw.mu.Unlock()
	unitsReady := 0
	for _, unit := range uw.Units {
		if unit.Ready {
			unitsReady++
		}
	}
	if unitsReady == len(units) {
		return "All systemd units are ready."
	} else {
		return fmt.Sprintf("Waiting for %d systemd units to be ready.", len(units)-unitsReady)
	}
}

// Returns number of units that aren't ready yet.
func (uw *UnitWatcher) NumUnitsNotReady() int {
	uw.mu.Lock()
	defer uw.mu.Unlock()

	unitsNotReady := len(uw.Units)
	for _, unit := range uw.Units {
		if unit.Ready {
			unitsNotReady--
		}
	}

	return unitsNotReady
}

func (uw *UnitWatcher) AllReady() bool {
	return uw.NumUnitsNotReady() == 0
}

func (uw *UnitWatcher) Elapsed() time.Duration {
	return time.Since(uw.started)
}

func (uw *UnitWatcher) NumUnits() int {
	return len(uw.Units)
}

func NewWatcher(dConn *dbus.Conn) *UnitWatcher {
	newUW := &UnitWatcher{
		DConn:   dConn,
		started: time.Now(),
	}

	setSystemdStartTime(dConn)

	newUW.Units = LoadUnitSpecs(config.UNIT_SPEC_PATH, newUW, SYSTEMD_START_TIME)

	return newUW
}

// Basic data for a unit's state.
type UnitInfo struct {
	Name            string
	SubStateDesired string // set to "any" if any substate is okay.
	Ready           bool   // observed substate matches desired substate
	LoadState       string
	ActiveState     string
	SubState        string
	Description     string // from systemd
	Desc            string // user-provided
	Properties      map[string]interface{}
	At              time.Time
	uw              *UnitWatcher
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
		timeChanged := getTimeOfStateChange(updates[1], properties)

		go func(obj *UnitStateChange, timeChanged, at time.Time, unitName string) {
			le := logging.NewLogEvent(fmt.Sprintf("%s state change.", unitName), obj)
			le.EndTime = timeChanged
			le.StartTime = at
			le.Duration = timeChanged.Sub(at)
			logging.Logs.AddLogEvent(le)
		}(obj, timeChanged, u.At, u.Name)

		u.At = timeChanged
		if config.SYSTEMD_ACCESS {
			u.Properties = properties
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
	return "SystemD unit state change"
}

func LoadUnitSpecs(filename string, uw *UnitWatcher, startTime time.Time) []*UnitInfo {
	type ParseObj struct {
		SpecArr []UnitSpec `json:"unitSpecs"`
	}

	temp := ParseObj{}
	units := make([]*UnitInfo, 0)
	if _, err := os.Stat(config.UNIT_SPEC_PATH); errors.Is(err, os.ErrNotExist) {
		return units
	}

	bytes, err := os.ReadFile(config.UNIT_SPEC_PATH)
	if err != nil {
		log.Fatal(fmt.Errorf("Reading unit specs: %s", err.Error()))
	}

	err = json.Unmarshal(bytes, &temp)
	if err != nil {
		log.Fatal(fmt.Errorf("Reading unit specs: %s", err.Error()))
	}

	for _, s := range temp.SpecArr {
		units = append(units, &UnitInfo{
			s.Name,
			s.SubStateDesired,
			false,
			"",
			"",
			"",
			"",
			s.Desc,
			nil,
			startTime,
			uw,
		})
	}

	return units
}

func assertString(obj interface{}) string {
	s, ok := obj.(string)
	if !ok {
		log.Fatal(errors.New(fmt.Sprintf("Type assertion failed: %v is a %T.", obj, obj)))
	}
	return s
}

func assertUint64(obj interface{}) uint64 {
	i, ok := obj.(uint64)
	if !ok {
		log.Fatal(errors.New(fmt.Sprintf("Type assertion failed: %v is a %T.", obj, obj)))
	}
	return i
}

func convertRealtime(val uint64) (int64, int64) { // may have to account for different levels of clock precision across machines
	var div uint64 = 1e6

	sec, nsec := int64(val/div), int64((val%div)*1000) // systemd timestamps are microsecond precision
	//log.Printf("sec: %d nsec: %d", sec, nsec)
	return sec, nsec
}

func getTimeOfStateChange(activeState string, properties map[string]interface{}) time.Time {
	var key string
	switch activeState {
	case "inactive", "failed":
		key = "InactiveEnterTimestamp"
	case "activating":
		key = "InactiveExitTimestamp"
	case "active":
		key = "ActiveEnterTimestamp"
	case "deactivating":
		key = "ActiveExitTimestamp"
	default:
		log.Fatalf("%s is an unrecognized active state.", activeState)
	}

	realTime := assertUint64(properties[key])

	if realTime == 0 { // state is the same as it was when it started.
		return SYSTEMD_START_TIME
	}
	sec, nsec := convertRealtime(realTime)
	return time.Unix(sec, nsec)
}

func setSystemdStartTime(dConn *dbus.Conn) {
	props, err := dConn.GetAllProperties("-.slice")
	if err != nil {
		log.Fatal(err)
	}

	realTime := assertUint64(props["ActiveEnterTimestamp"])
	sec, nsec := convertRealtime(realTime)
	SYSTEMD_START_TIME = time.Unix(sec, nsec)

	// log when systemd was started
	msg := "SystemD was started."
	msgLog := logging.NewLogEvent(msg, &logging.MessageLog{msg, "SystemD"})

	msgLog.StartTime = SYSTEMD_START_TIME
	msgLog.EndTime = SYSTEMD_START_TIME
	msgLog.Duration = SYSTEMD_START_TIME.Sub(SYSTEMD_START_TIME)

	logging.Logs.AddLogEvent(msgLog)
}
