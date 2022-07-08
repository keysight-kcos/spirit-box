package scripts

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func RunScripts() {
	runScriptsInDir()
	loadScriptList()
}

func runScriptsInDir() {
	scriptDir := "/usr/share/spirit-box/"
	items, _ := ioutil.ReadDir(scriptDir)
	fmt.Println("scripting...")
	for _, item := range items {
		if !item.IsDir() && item.Name()[len(item.Name())-3:] == ".sh" {
			fmt.Println("running script " + scriptDir + item.Name())
			out, err := exec.Command("/bin/sh", scriptDir+item.Name()).Output()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s", out)
		}
	}
}

func loadScriptList() ([]string, error) {
	var lines []string
	path := "/usr/share/spirit-box/scripts"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		fmt.Println("No script file.")
		return lines, err
	}
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := os.Stat(line); errors.Is(err, os.ErrNotExist) {
			log.Fatal(errors.New("Script does not exist: " + line))
		} else {
			fmt.Println("running script " + line)
			lines = append(lines, line)
			out, err := exec.Command("/bin/sh", line).Output()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s", out)
		}

	}
	return lines, scanner.Err()
}
