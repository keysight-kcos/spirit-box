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

func RunAllScripts() {
	runScriptsInDir()
	loadScriptList()
}

func runScriptsInDir(){
	l := logging.Logger
	scriptDir := "/usr/share/spirit-box/"
	items, _ := ioutil.ReadDir(scriptDir)
	fmt.Printf("Running scripts in %s\n", scriptDir);
	for _, item := range items {
		if !item.IsDir() && item.Name()[len(item.Name())-3:] == ".sh"{
			fmt.Println("Running script " + scriptDir + item.Name() + "...")
			out, err := exec.Command("/bin/sh", scriptDir + item.Name()).Output()
			if err != nil{
				log.Fatal(err)
			}
			fmt.Printf("%s", out)
			l.Printf("Ran %s%s", scriptDir, item.Name())
		}
	}
	fmt.Println()
}

func loadScriptList() ([]string, error) {
	var lines []string
	path := "/usr/share/spirit-box/scripts"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist){
		fmt.Println("No script file.")
		return lines, err
	}

	fmt.Printf("Running scripts based on path names in %s\n", path)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	l := logging.Logger
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := os.Stat(line); errors.Is(err, os.ErrNotExist) {
			log.Fatal(errors.New("Script does not exist: " + line))
		} else {
			fmt.Println("Running script " + line + "...")
			lines = append(lines, line)
			out, err := exec.Command("/bin/sh", line).Output()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s", out)
			l.Printf("Ran %s", line)
		}

	}
	fmt.Println()
	return lines, scanner.Err()
}