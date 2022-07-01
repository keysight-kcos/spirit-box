package scripts

import(
	"fmt"
	"log"
	"io/ioutil"
	"os/exec"
	"os"
	"errors"
	"bufio"
	"spirit-box/logging"
	"encoding/json"
	"sort"
)

type ScriptData struct{
	Path string
	Shell string
	Priority int
	Output string
	Pid int
	StartTime int
	EndTime int
	Exitcode int
}

type ByPriority []ScriptData
func (d ByPriority) Len() int {return len(d)}
func (d ByPriority) Less(i, j int) bool {return d[i].Priority < d[j].Priority}
func (d ByPriority) Swap(i, j int) {d[i], d[j] = d[j], d[i]}

func RunAllScripts() {
	l := logging.Logger
	runScriptsInDir()
	scriptList, _ := loadScriptJson(l)
	scriptList = sanitizeScriptList(scriptList)
	lists := organizeLists(scriptList)
	//scriptList, _ := loadScriptList()
	runScriptList(l, lists)
}

func checkShebang(line string) (bool, string){
/*checks if the first 2 characters of a file are shebang
inputs: string - the file path
outputs: bool - true if shebang exists
         string - the path of the shell to use*/
	file, err := os.Open(line)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	shebang := scanner.Text()
	if len(shebang) < 2 || shebang[:2] != "#!"{
		return false, ""
	}
	shell := shebang[2:]
	return true, shell
}

func executeAndChan(l *log.Logger, scriptData ScriptData, co chan<- ScriptData) {
/*executes a script
inputs: *log.Logger - logger
        ScriptData - data including path to script and shell
	chan<-ScriptData - channel to collect goroutine output ScriptData*/
	fmt.Println("Running script " + scriptData.Path + "...")
	out, err := exec.Command(scriptData.Shell, scriptData.Path).Output()
	if err != nil {
		log.Fatal(err)
	}
	scriptData.Output = string(out)
	co <- scriptData
	l.Printf("Ran %s", scriptData.Path)
}

func runScriptsInDir(){
/*runs the scripts in hard coded directory*/
	l := logging.Logger
	outputChannel := make(chan ScriptData)
	scriptData := ScriptData{}
	scriptCount := 0
	scriptDir := "/usr/share/spirit-box/"
	items, _ := ioutil.ReadDir(scriptDir)
	fmt.Printf("Running scripts in %s\n", scriptDir);
	for _, item := range items {
		if !item.IsDir() {
			isScript, shell := checkShebang(scriptDir+item.Name())
			if isScript {
				scriptData.Shell = shell
				scriptData.Path = scriptDir+item.Name()
				scriptCount++
				go executeAndChan(l, scriptData, outputChannel)
			}
		}
	}
	for i := 0; i<scriptCount; i++{
		fmt.Print((<-outputChannel).Output)
	}
	fmt.Println()
}

func runScriptList(l *log.Logger, scriptList [][]ScriptData) {
/*runs scripts in array
inputs: *log.Logger - log
	[][]ScriptData - list of lists of scripts to run*/
	outputChannel := make(chan ScriptData)
	for i:= 0; i<len(scriptList); i++{
		fmt.Printf("Running scripts of priority %d\n", scriptList[i][0].Priority)
		for j:= 0; j<len(scriptList[i]); j++{
			go executeAndChan(l, scriptList[i][j], outputChannel)
		}
		for j := 0; j<len(scriptList[i]); j++{
			fmt.Print((<-outputChannel).Output)
		}
	}
}

func organizeLists(scriptList []ScriptData)([][]ScriptData){
/*organizes scriptData into list of lists by priority
inputs: []ScriptData - list of ScriptData ordered by priority
outputs: [][]ScriptData - list of lists of ScriptData*/
	var lists [][]ScriptData
	if len(scriptList) < 1{
		return lists
	}
	index := 0
	currPrio := scriptList[0].Priority
	lists = append(lists, []ScriptData{})
	for i := 0; i<len(scriptList); i++{
		scriptData := scriptList[i]
		if scriptData.Priority > currPrio{
			index++
			currPrio = scriptData.Priority
			lists = append(lists, []ScriptData{})
		}
		lists[index] = append(lists[index], scriptData)
	}
	return lists
}

func sanitizeScriptList(scriptList []ScriptData) ([]ScriptData){
/*sanitizes scripts in list
inputs: []ScriptData - list of scripts
outputs: []ScriptData - sanitized list*/
	var sanitized []ScriptData
	for i := 0; i<len(scriptList); i++{
		scriptData := scriptList[i]
		if _, err := os.Stat(scriptData.Path); errors.Is(err, os.ErrNotExist) {
			log.Fatal(errors.New("Script does not exist: " + scriptData.Path))
		} else if isScript, shell := checkShebang(scriptData.Path); !isScript {
			fmt.Printf("Not shebang: %s\n", scriptData.Path);
		} else {
			scriptData.Shell = shell
			sanitized = append(sanitized, scriptData)
		}
	}
	sort.Sort(ByPriority(sanitized))
	return sanitized
}

func loadScriptJson(l *log.Logger) ([]ScriptData, error) {
/*loads scriptData listed in scripts.json file
outputs: []ScriptData - array of scripts
         error - errors*/
	path := "/usr/share/spirit-box/scripts.json"
	fmt.Printf("Loading scripts based on path names in %s\n", path)
	var scriptList []ScriptData
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist){
		fmt.Println("No script json.")
		return scriptList, err
	}
	content, err := ioutil.ReadFile(path)
	if err != nil{
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &scriptList)
	if err != nil {
		fmt.Println(err)
		l.Printf("Error in scripts json: %s", err)
	}
	return scriptList, err
}

func loadScriptList() ([]ScriptData, error) {
/*executes scripts listed as paths in script file
outputs: []ScriptData - array of scripts
         error - errors*/
	var scriptList []ScriptData
	path := "/usr/share/spirit-box/scripts"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist){
		fmt.Println("No script file.")
		return scriptList, err
	}

	fmt.Printf("Running scripts based on path names in %s\n", path)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scriptData := ScriptData{}
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := os.Stat(line); errors.Is(err, os.ErrNotExist) {
			log.Fatal(errors.New("Script does not exist: " + line))
		} else if isScript, shell := checkShebang(line); !isScript {
			fmt.Printf("Not shebang: %s\n", line);
		} else {
			scriptData.Path = line
			scriptData.Shell = shell
			scriptList = append(scriptList, scriptData)
		}
	}
	return scriptList, scanner.Err()
}
