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
)

type ScriptData struct{
	PID int
	path string
	shell string
	startTime int
	endTime int
	output string
	exitcode int
	priority int
}

func RunAllScripts() {
	runScriptsInDir()
	loadScriptList()
}

func checkShebang(line string) (bool, string){
/*checks if the first 2 characters of a file are shebang
inputs: string - the file path
outputs: bool - true if shebang exists
         string - the path of the shell to use*/
	isScript := true
	file, err := os.Open(line)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	shebang := scanner.Text()
	if len(shebang) >= 2 && shebang[:2] != "#!"{
		return false, ""
	}
	shell := shebang[2:]
	return isScript, shell
}

func executeAndOutput(l *log.Logger, scriptData ScriptData, co chan<- ScriptData) {
/*executes a script
inputs: *log.Logger - logger
        ScriptData - data including path to script
	chan<-ScriptData - channel to collect goroutine output ScriptData*/
	fmt.Println("Running script " + scriptData.path + "...")
	out, err := exec.Command(scriptData.shell, scriptData.path).Output()
		if err != nil {
		log.Fatal(err)
	}
	scriptData.output = string(out)
	co <- scriptData
	l.Printf("Ran %s", scriptData.path)
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
				scriptData.shell = shell
				scriptData.path = scriptDir+item.Name()
				scriptCount++
				go executeAndOutput(l, scriptData, outputChannel)
			}
		}
	}
	for i := 0; i<scriptCount; i++{
		fmt.Print((<-outputChannel).output)
	}
	fmt.Println()
}

func runScriptList(l *log.Logger, scriptList []ScriptData) {
/*runs scripts in array
inputs: *log.Logger - log
	[]ScriptData - list of scripts to run*/
	outputChannel := make(chan ScriptData)

	for i:= 0; i<len(scriptList); i++{
		go executeAndOutput(l, scriptList[i], outputChannel)
	}
	for i := 0; i<len(scriptList); i++{
		fmt.Print((<-outputChannel).output)
	}
}

func loadScriptList() ([]ScriptData, error) {
/*executes scripts listed as paths in script file
outputs: []string - array of paths it attempts to execute
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

	l := logging.Logger
	scanner := bufio.NewScanner(file)
	scriptData := ScriptData{}
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := os.Stat(line); errors.Is(err, os.ErrNotExist) {
			log.Fatal(errors.New("Script does not exist: " + line))
		} else if isScript, shell := checkShebang(line); !isScript {
			fmt.Printf("Not shebang: %s\n", scriptData.path);
		} else {
			scriptData.path = line
			scriptData.shell = shell
			scriptList = append(scriptList, scriptData)
		}
	}
	runScriptList(l, scriptList)
	return scriptList, scanner.Err()
}
