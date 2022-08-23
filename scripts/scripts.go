package scripts

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"spirit-box/config"
	"spirit-box/logging"
	"strings"
	"sync"
	"time"
)

// Loaded from a json file.
// Specifications for how to run a script.
type ScriptSpec struct {
	Cmd           string   `json:"cmd"`
	Args          []string `json:"args"`
	Priority      int      `json:"priority"`
	RetryTimeout  int      `json:"retryTimeout"`  // time in ms between retrying a failed script
	TotalWaitTime int      `json:"totalWaitTime"` // the maximum amount of time in ms to wait for a success
}

func (s *ScriptSpec) Run() ScriptResult {
	cmd := exec.Command(s.Cmd, s.Args...)
	start := time.Now()
	bytes, err := cmd.Output()
	elapsed := time.Since(start)
	if err != nil {
		log.Fatal(err)
	}

	res := ScriptResult{}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		log.Fatal(err)
	}
	res.Pid = cmd.Process.Pid
	res.StartTime = start
	res.ElapsedTime = elapsed

	return res
}

func (s *ScriptSpec) ToString() string {
	return fmt.Sprintf("%s %s", s.Cmd, strings.Join(s.Args, " "))
}

type ScriptResult struct {
	Success     bool          `json:"success"`
	Info        string        `json:"info"` // More detailed information the script may want to return.
	Pid         int           `json:"pid"`
	StartTime   time.Time     `json:"startTime"`
	ElapsedTime time.Duration `json:"elaspedTime[ns]"`
}

type ScriptTracker struct {
	StartTime time.Time       `json:"startTime"`
	EndTime   time.Time       `json:"endTime"`
	Runs      []*ScriptResult `json:"runs"` // can add individual stats about each run later
	Finished  bool            `json:"finished"`
}

func (st *ScriptTracker) ToString() string {
	var runs string
	for _, res := range st.Runs {
		runs += fmt.Sprintf("%v\n", *res)
	}
	return fmt.Sprintf(
		"Start: %s\nEnd: %s\nRuns:\n%s",
		st.StartTime,
		st.EndTime,
		runs,
	)
}

func (st *ScriptTracker) Succeeded() bool {
	return len(st.Runs) > 0 && st.Runs[len(st.Runs)-1].Success
}

type ScriptLogObj struct { // for json logs
	StartTime time.Time       `json:"-"`
	EndTime   time.Time       `json:"-"`
	Runs      []*ScriptResult `json:"runs"`
	Spec      *ScriptSpec     `json:"scriptSpecification"`
	Succeeded bool            `json:"succeeded"`
	Name      string          `json:"name"`
}

func NewScriptLogObj(spec *ScriptSpec, tracker *ScriptTracker) *ScriptLogObj {
	return &ScriptLogObj{
		StartTime: tracker.StartTime,
		EndTime:   tracker.EndTime,
		Runs:      tracker.Runs,
		Spec:      spec,
		Succeeded: tracker.Succeeded(),
		Name:      strings.Split(spec.Cmd, "/")[len(strings.Split(spec.Cmd, "/"))-1],
	}
}

func (sl *ScriptLogObj) LogLine() string {
	return fmt.Sprintf("Executed '%s' %d times. Success: %t", sl.Spec.ToString(), len(sl.Runs), sl.Succeeded)
}

func (sl *ScriptLogObj) GetObjType() string {
	return "Script event."
}

type PriorityGroup struct {
	Num      int              `json:"num"`
	Specs    []*ScriptSpec    `json:"specs"`
	Trackers []*ScriptTracker `json:"trackers"`
}

func (pg *PriorityGroup) RunAll() {
	// Init trackers
	now := time.Now()
	pg.Trackers = make([]*ScriptTracker, len(pg.Specs))
	for i, _ := range pg.Specs {
		pg.Trackers[i] = &ScriptTracker{
			StartTime: now,
			Runs:      make([]*ScriptResult, 0, 1000),
		}
	}

	var wg sync.WaitGroup
	for i, _ := range pg.Specs {
		wg.Add(1)
		go func(spec *ScriptSpec, tracker *ScriptTracker) {
			timer := time.NewTimer(time.Duration(spec.TotalWaitTime) * time.Millisecond)
			resChan := make(chan ScriptResult)
		RLoop:
			for {
				go func() {
					resChan <- spec.Run()
				}()
				select {
				case res := <-resChan:
					tracker.Runs = append(tracker.Runs, &res)
					if res.Success {
						break RLoop
					}
				case <-timer.C: // process took too long
					break RLoop
				}
				time.Sleep(time.Duration(spec.RetryTimeout) * time.Millisecond)
			}

			tracker.EndTime = time.Now()
			tracker.Finished = true

			go func(spec *ScriptSpec, tracker *ScriptTracker) { // logging
				scriptLog := NewScriptLogObj(spec, tracker)
				le := logging.NewLogEvent(scriptLog.LogLine(), scriptLog)
				le.StartTime = scriptLog.StartTime
				le.EndTime = scriptLog.EndTime
				le.Duration = scriptLog.EndTime.Sub(scriptLog.StartTime)
				logging.Logs.AddLogEvent(le)
			}(spec, tracker)

			wg.Done()
		}(pg.Specs[i], pg.Trackers[i])
	}
	wg.Wait()
}

func (pg *PriorityGroup) AllSucceeded() bool {
	if pg.Trackers == nil {
		return false
	}

	all := true
	for _, tracker := range pg.Trackers {
		all = all && tracker.Succeeded()
	}
	return all
}

func (pg *PriorityGroup) GetStatus() (int, int) {
	running := 0
	numFailed := 0

	for _, tracker := range pg.Trackers {
		if !tracker.Finished {
			running += 1
			continue
		}
		if !tracker.Succeeded() {
			numFailed += 1
		}
	}

	return running, numFailed
}

func (pg *PriorityGroup) GetLongestCmdLength() int { // for formatting in tui
	max := 0
	for _, spec := range pg.Specs {
		length := len(spec.Cmd)
		for _, arg := range spec.Args {
			length += len(arg) + 1
		}

		if length > max {
			max = length
		}
	}
	return max
}

func (pg *PriorityGroup) PrintAfterRun() { // Print the results of a run. For debugging.
	fmt.Printf("Priority group %d:\n", pg.Num)
	for i, s := range pg.Specs {
		fmt.Printf("%s %v:\n%s\n", s.Cmd, s.Args, pg.Trackers[i].ToString())
	}
}

type ScriptController struct {
	PriorityGroups []*PriorityGroup `json:"priorityGroups"`
	NumScripts     int              `json:"-"`
}

func (sc *ScriptController) RunPriorityGroups() {
	for _, pg := range sc.PriorityGroups {
		//fmt.Printf("Running scripts in priority group %d:\n", pg.num)
		pg.RunAll()
		//pg.PrintAfterRun()
	}
}

func (sc *ScriptController) GetLongestCmdLength() int { // for formatting in tui
	max := 0
	for _, pg := range sc.PriorityGroups {
		length := pg.GetLongestCmdLength()
		if length > max {
			max = length
		}
	}
	return max
}

func (sc *ScriptController) PrintPriorityGroups() { // for debugging
	// output should be ordered by priority group
	for _, pg := range sc.PriorityGroups {
		fmt.Printf("PriorityGroup %d:\n", pg.Num)
		for _, s := range pg.Specs {
			fmt.Println(*s)
		}
	}
}

func (sc *ScriptController) AllReady() bool {
	allReady := true
	for _, pg := range sc.PriorityGroups {
		running, numFailed := pg.GetStatus()
		if running > 0 || numFailed > 0 {
			allReady = false
			break
		}
	}

	return allReady
}

func (sc *ScriptController) GetStatus() (int, int) {
	running := 0
	failed := 0
	for _, pg := range sc.PriorityGroups {
		r, f := pg.GetStatus()
		running += r
		failed += f
	}
	return running, failed
}

func NewController() *ScriptController {
	priorities := make(map[int]PriorityGroup)
	specs := LoadScriptSpecs()
	maxPriority := 0 // Assuming negative priorities are not a thing.
	numPGroups := 0

	for _, temp := range specs {
		s := temp
		if _, ok := priorities[s.Priority]; !ok {
			pg := PriorityGroup{Num: s.Priority}
			pg.Specs = make([]*ScriptSpec, 1)
			pg.Specs[0] = &s
			priorities[s.Priority] = pg
			numPGroups++
		} else {
			newSpecs := append(
				priorities[s.Priority].Specs,
				&s,
			)
			pg := PriorityGroup{
				Num:   s.Priority,
				Specs: newSpecs,
			}
			priorities[s.Priority] = pg
		}

		if s.Priority > maxPriority {
			maxPriority = s.Priority
		}
	}

	sc := &ScriptController{
		PriorityGroups: make([]*PriorityGroup, numPGroups),
		NumScripts:     len(specs),
	}

	counter := 0
	for i := 0; i < maxPriority+1; i++ { // inefficient, can optimize later
		if pg, ok := priorities[i]; ok {
			sc.PriorityGroups[counter] = &pg
			counter++
		}
	}

	return sc
}

func LoadScriptSpecs() []ScriptSpec {
	type ParseObj struct {
		SpecArr []ScriptSpec `json:"scriptSpecs"`
	}

	temp := ParseObj{}
	specs := make([]ScriptSpec, 0)
	if _, err := os.Stat(config.SCRIPT_SPEC_PATH); errors.Is(err, os.ErrNotExist) {
		return specs
	}

	bytes, err := os.ReadFile(config.SCRIPT_SPEC_PATH)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &temp)
	if err != nil {
		log.Fatal(err)
	}

	return temp.SpecArr
}

// used in TUI
type ScriptStatus struct {
	Cmd    string
	Status int // 0: waiting 1: running 2: failed, 3: succeeded
}

// just get statuses of individual scripts for displaying in the top level.
func (sc *ScriptController) GetScriptStatuses() []ScriptStatus {
	ret := make([]ScriptStatus, 0, sc.NumScripts)

	for _, pg := range sc.PriorityGroups {
		for j, spec := range pg.Specs {
			cmdStr := spec.ToString()
			stat := 0

			if pg.Trackers != nil {
				tracker := pg.Trackers[j]
				if tracker.Finished {
					if tracker.Succeeded() {
						stat = 3
					} else {
						stat = 2
					}
				} else {
					stat = 1
				}
			}

			ret = append(ret, ScriptStatus{Cmd: cmdStr, Status: stat})
		}
	}

	return ret
}
