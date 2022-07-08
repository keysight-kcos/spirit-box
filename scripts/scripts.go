package scripts

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

const SCRIPT_SPEC_PATH = "/usr/share/spirit-box/script_specs.json"

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
	bytes, err := exec.Command(s.Cmd, s.Args...).Output()
	if err != nil {
		log.Fatal(err)
	}

	res := ScriptResult{}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		log.Fatal(err)
	}

	return res
}

type ScriptResult struct {
	Success bool   `json:"success"`
	Info    string `json:"info"` // More detailed information the script may want to return.
}

type ScriptTracker struct {
	StartTime time.Time
	EndTime   time.Time
	Runs      []*ScriptResult // can add individual stats about each run later
	Finished  bool
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

type PriorityGroup struct {
	Num      int
	Specs    []*ScriptSpec
	Trackers []*ScriptTracker
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
	for i, s := range pg.Specs {
		wg.Add(1)
		go func(index int, spec *ScriptSpec) {
			timer := time.NewTimer(time.Duration(spec.TotalWaitTime) * time.Millisecond)
			resChan := make(chan ScriptResult)
		RLoop:
			for {
				go func() {
					resChan <- spec.Run()
				}()
				select {
				case res := <-resChan:
					pg.Trackers[index].Runs = append(pg.Trackers[index].Runs, &res)
					if res.Success {
						break RLoop
					}
				case <-timer.C: // process took too long
					break RLoop
				}
				time.Sleep(time.Duration(spec.RetryTimeout) * time.Millisecond)
			}
			pg.Trackers[index].EndTime = time.Now()
			pg.Trackers[index].Finished = true
			wg.Done()
		}(i, s)
	}
	wg.Wait()
}

func (pg *PriorityGroup) PrintAfterRun() { // Print the results of a run. For debugging.
	fmt.Printf("Priority group %d:\n", pg.Num)
	for i, s := range pg.Specs {
		fmt.Printf("%s %v:\n%s\n", s.Cmd, s.Args, pg.Trackers[i].ToString())
	}
}

type ScriptController struct {
	PriorityGroups []*PriorityGroup
}

func (sc *ScriptController) RunPriorityGroups() {
	for _, pg := range sc.PriorityGroups {
		//fmt.Printf("Running scripts in priority group %d:\n", pg.num)
		pg.RunAll()
		//pg.PrintAfterRun()
	}
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
	if _, err := os.Stat(SCRIPT_SPEC_PATH); errors.Is(err, os.ErrNotExist) {
		return specs
	}

	bytes, err := os.ReadFile(SCRIPT_SPEC_PATH)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &temp)
	if err != nil {
		log.Fatal(err)
	}

	return temp.SpecArr
}
